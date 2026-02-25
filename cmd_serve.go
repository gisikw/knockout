package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// subscriber represents a client subscribed to SSE updates for a project.
type subscriber struct {
	project string
	ch      chan string
}

// tailer manages the global event stream and broadcasts to subscribers.
type tailer struct {
	mu          sync.Mutex
	subscribers map[*subscriber]bool
	started     bool
	eventID     int64
}

var globalTailer = &tailer{
	subscribers: make(map[*subscriber]bool),
}

// subscribe registers a new subscriber for a project.
func (t *tailer) subscribe(project string, ch chan string) *subscriber {
	t.mu.Lock()
	defer t.mu.Unlock()
	sub := &subscriber{project: project, ch: ch}
	t.subscribers[sub] = true
	return sub
}

// unsubscribe removes a subscriber.
func (t *tailer) unsubscribe(sub *subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.subscribers, sub)
	close(sub.ch)
}

// start launches the event stream tailer goroutine (idempotent).
func (t *tailer) start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.started {
		return
	}
	t.started = true
	go t.tailEventStream()
}

// broadcast sends an SSE message to all subscribers for a given project.
func (t *tailer) broadcast(project, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for sub := range t.subscribers {
		if sub.project == project {
			select {
			case sub.ch <- message:
			default:
				// Channel full, drop event (best-effort delivery)
			}
		}
	}
}

// nextEventID returns the next monotonic event ID.
func (t *tailer) nextEventID() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.eventID++
	return t.eventID
}

// tailEventStream watches the mutation event file and broadcasts to subscribers.
func (t *tailer) tailEventStream() {
	path := mutationEventPath()
	if path == "" {
		return
	}

	for {
		if err := t.tailFile(path); err != nil {
			time.Sleep(5 * time.Second)
		}
	}
}

// tailFile reads from the event stream file, handling EOF and retries.
func (t *tailer) tailFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Wait for file to exist
			for {
				time.Sleep(2 * time.Second)
				f, err = os.Open(path)
				if err == nil {
					break
				}
				if !os.IsNotExist(err) {
					return err
				}
			}
		} else {
			return err
		}
	}
	defer f.Close()

	// Seek to end — only process new events
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	buf := make([]byte, 4096)
	var partial []byte

	for {
		n, err := f.Read(buf)
		if n > 0 {
			partial = append(partial, buf[:n]...)
			// Process complete lines
			for {
				newline := -1
				for i, b := range partial {
					if b == '\n' {
						newline = i
						break
					}
				}
				if newline < 0 {
					break
				}
				line := partial[:newline]
				partial = partial[newline+1:]

				if len(line) == 0 {
					continue
				}

				var evt MutationEvent
				if err := json.Unmarshal(line, &evt); err != nil {
					continue
				}

				// Re-query the project and broadcast to subscribers
				t.broadcastToProject(evt.Project)
			}
		}
		if err != nil {
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}
	}
}

// broadcastToProject re-queries a project and sends the result as SSE to all subscribers.
func (t *tailer) broadcastToProject(projectPath string) {
	// Find tickets directory for this project
	ticketsDir := projectPath + "/.ko/tickets"

	// Query tickets directly (avoid exec for testability)
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return // Best-effort: skip on error
	}
	SortByPriorityThenModified(tickets)

	// Format as SSE event
	eventID := t.nextEventID()
	var message strings.Builder
	fmt.Fprintf(&message, "id: %d\n", eventID)

	// Serialize tickets as JSONL with "data: " prefix
	for _, ticket := range tickets {
		j := ticketToJSON(ticket, ticketsDir)
		line, err := json.Marshal(j)
		if err != nil {
			continue
		}
		fmt.Fprintf(&message, "data: %s\n", string(line))
	}
	fmt.Fprintf(&message, "\n")

	t.broadcast(projectPath, message.String())
}

// handleSubscribe handles SSE subscriptions for a project using the global tailer.
func handleSubscribe(w http.ResponseWriter, r *http.Request) {
	handleSubscribeWithTailer(w, r, globalTailer)
}

// handleSubscribeWithTailer handles SSE subscriptions for a project with a specific tailer.
func handleSubscribeWithTailer(w http.ResponseWriter, r *http.Request, t *tailer) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract project from URL path or query parameter
	// Path format: /subscribe/{project} where project is a #tag
	// Query format: /subscribe?project=/absolute/path
	var projectPath string

	if q := r.URL.Query().Get("project"); q != "" {
		projectPath = q
	} else {
		path := strings.TrimPrefix(r.URL.Path, "/subscribe/")
		if path == "" || path == r.URL.Path {
			http.Error(w, "project parameter required", http.StatusBadRequest)
			return
		}

		if strings.HasPrefix(path, "#") {
			// Registry lookup for #tag syntax
			regPath := RegistryPath()
			if regPath == "" {
				http.Error(w, "cannot determine config directory", http.StatusInternalServerError)
				return
			}
			reg, err := LoadRegistry(regPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("registry error: %v", err), http.StatusInternalServerError)
				return
			}
			tag := strings.TrimPrefix(path, "#")
			projPath, ok := reg.Projects[tag]
			if !ok {
				http.Error(w, fmt.Sprintf("project not found: %s", path), http.StatusNotFound)
				return
			}
			projectPath = projPath
		} else {
			http.Error(w, "use #tag syntax or ?project= query param", http.StatusBadRequest)
			return
		}
	}

	// Get initial snapshot
	ticketsDir := projectPath + "/.ko/tickets"
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}
	SortByPriorityThenModified(tickets)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send retry directive
	fmt.Fprintf(w, "retry: 3000\n\n")
	flusher.Flush()

	// Send initial snapshot with id=0
	fmt.Fprintf(w, "id: 0\n")
	for _, ticket := range tickets {
		j := ticketToJSON(ticket, ticketsDir)
		line, jsonErr := json.Marshal(j)
		if jsonErr != nil {
			continue
		}
		fmt.Fprintf(w, "data: %s\n", string(line))
	}
	fmt.Fprintf(w, "\n")
	flusher.Flush()

	// Register subscriber
	ch := make(chan string, 10) // Buffered for backpressure handling
	sub := t.subscribe(projectPath, ch)
	defer t.unsubscribe(sub)

	// Block and forward events until client disconnects
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}

func cmdServe(args []string) int {
	// Parse flags
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	port := fs.String("p", "9876", "port to listen on")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Define whitelist of allowed subcommands
	whitelist := map[string]bool{
		"ls":        true,
		"ready":     true,
		"blocked":   true,
		"resolved":  true,
		"closed":    true,
		"query":     true,
		"show":      true,
		"questions": true,
		"answer":    true,
		"close":     true,
		"reopen":    true,
		"block":     true,
		"start":     true,
		"bump":      true,
		"note":      true,
		"status":    true,
		"dep":       true,
		"undep":     true,
		"agent":     true,
	}

	// Start global event tailer
	globalTailer.start()

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe/", handleSubscribe)
	mux.HandleFunc("/ko", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var req struct {
			Argv    []string `json:"argv"`
			Project string   `json:"project"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Validate argv
		if len(req.Argv) == 0 {
			http.Error(w, "argv must have at least one element", http.StatusBadRequest)
			return
		}

		// Check if first element is whitelisted
		subcommand := req.Argv[0]
		if !whitelist[subcommand] {
			errResp := map[string]string{
				"error": fmt.Sprintf("subcommand '%s' not allowed", subcommand),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		// Resolve project path if specified
		var projectPath string
		if req.Project != "" {
			if strings.HasPrefix(req.Project, "#") {
				// Registry lookup for #tag syntax
				regPath := RegistryPath()
				if regPath == "" {
					http.Error(w, "cannot determine config directory", http.StatusInternalServerError)
					return
				}
				reg, err := LoadRegistry(regPath)
				if err != nil {
					errResp := map[string]string{
						"error": fmt.Sprintf("registry error: %v", err),
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(errResp)
					return
				}
				tag := strings.TrimPrefix(req.Project, "#")
				path, ok := reg.Projects[tag]
				if !ok {
					errResp := map[string]string{
						"error": fmt.Sprintf("project not found: %s", req.Project),
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(errResp)
					return
				}
				projectPath = path
			} else {
				// Treat as absolute path
				projectPath = req.Project
			}
		}

		// Execute command
		cmd := exec.Command(os.Args[0], req.Argv...)
		if projectPath != "" {
			cmd.Dir = projectPath
		}
		output, err := cmd.CombinedOutput()

		if err != nil {
			// Non-zero exit code: return 400 with stderr
			errResp := map[string]string{
				"error": string(output),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		// Success: return 200 with stdout
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	})

	// Create server
	addr := ":" + *port
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel to signal server started
	serverStarted := make(chan struct{})

	// Start server in goroutine
	go func() {
		fmt.Fprintf(os.Stdout, "ko serve: listening on %s\n", addr)
		close(serverStarted)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "ko serve: %v\n", err)
		}
	}()

	// Wait for server to start
	<-serverStarted

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// Wait for signal
	sig := <-sigCh
	fmt.Fprintf(os.Stdout, "ko serve: received %v, shutting down\n", sig)

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ko serve: shutdown error: %v\n", err)
		return 1
	}

	return 0
}
