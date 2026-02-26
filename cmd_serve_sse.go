package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// subscriber represents a client subscribed to SSE updates for a project.
type subscriber struct {
	project      string
	ch           chan string
	includeAgent bool // true for /status/ endpoint, false for /subscribe/
}

// tailer manages the global event stream and broadcasts to subscribers.
type tailer struct {
	mu              sync.Mutex
	subscribers     map[*subscriber]bool
	started         bool
	eventID         int64
	agentPollQuit   chan struct{}
	agentPollStatus map[string]agentStatusJSON // tracks last known status per project
}

var globalTailer = &tailer{
	subscribers:     make(map[*subscriber]bool),
	agentPollStatus: make(map[string]agentStatusJSON),
}

// subscribe registers a new subscriber for a project.
func (t *tailer) subscribe(project string, ch chan string, includeAgent bool) *subscriber {
	t.mu.Lock()
	defer t.mu.Unlock()
	sub := &subscriber{project: project, ch: ch, includeAgent: includeAgent}
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
	t.agentPollQuit = make(chan struct{})
	go t.tailEventStream()
	go t.pollAgentStatus()
}

// broadcast sends an SSE message to all subscribers for a given project (tickets-only subscribers).
func (t *tailer) broadcast(project, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for sub := range t.subscribers {
		if sub.project == project && !sub.includeAgent {
			select {
			case sub.ch <- message:
			default:
				// Channel full, drop event (best-effort delivery)
			}
		}
	}
}

// broadcastToSubscribers sends an SSE message to subscribers matching the filter.
func (t *tailer) broadcastToSubscribers(project, message string, includeAgentOnly bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for sub := range t.subscribers {
		if sub.project == project && (!includeAgentOnly || sub.includeAgent) {
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

	// Format as SSE event for /subscribe/ endpoint (tickets only, no type wrapper)
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

	// Also send to /status/ subscribers with type wrappers
	eventID2 := t.nextEventID()
	var statusMessage strings.Builder
	fmt.Fprintf(&statusMessage, "id: %d\n", eventID2)

	for _, ticket := range tickets {
		j := ticketToJSON(ticket, ticketsDir)
		ticketEnvelope := map[string]interface{}{
			"type":   "ticket",
			"ticket": j,
		}
		line, err := json.Marshal(ticketEnvelope)
		if err != nil {
			continue
		}
		fmt.Fprintf(&statusMessage, "data: %s\n", string(line))
	}
	fmt.Fprintf(&statusMessage, "\n")

	t.broadcastToSubscribers(projectPath, statusMessage.String(), true)
}

// getAgentStatus returns the agent status for the given tickets directory.
// This mirrors the logic in cmdAgentStatus but returns the struct directly.
func getAgentStatus(ticketsDir string) agentStatusJSON {
	status := agentStatusJSON{}

	// Check if pipeline config exists
	if _, err := FindPipelineConfig(ticketsDir); err != nil {
		return status
	}
	status.Provisioned = true

	pidPath := agentPidPath(ticketsDir)
	pid, err := readAgentPid(pidPath)
	if err != nil {
		// No PID file — check if a lock is held (orphaned agent)
		if isAgentLocked(ticketsDir) {
			status.Running = true
		}
		return status
	}

	if isProcessAlive(pid) {
		status.Running = true
		status.Pid = pid
		logPath := agentLogPath(ticketsDir)
		if last := lastLogLine(logPath); last != "" {
			status.LastLog = last
		}
	}
	return status
}

// pollAgentStatus polls agent status for all active projects every 2 seconds.
func (t *tailer) pollAgentStatus() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.agentPollQuit:
			return
		case <-ticker.C:
			t.checkAgentStatusChanges()
		}
	}
}

// checkAgentStatusChanges checks agent status for all subscribed projects and broadcasts changes.
func (t *tailer) checkAgentStatusChanges() {
	t.mu.Lock()
	// Gather unique project paths from subscribers
	projects := make(map[string]bool)
	for sub := range t.subscribers {
		projects[sub.project] = true
	}
	t.mu.Unlock()

	// Check status for each project
	for projectPath := range projects {
		ticketsDir := projectPath + "/.ko/tickets"
		newStatus := getAgentStatus(ticketsDir)

		t.mu.Lock()
		oldStatus, exists := t.agentPollStatus[projectPath]
		changed := !exists || oldStatus != newStatus
		if changed {
			t.agentPollStatus[projectPath] = newStatus
		}
		t.mu.Unlock()

		// Broadcast if changed
		if changed {
			t.broadcastAgentStatus(projectPath, newStatus)
		}
	}
}

// broadcastAgentStatus sends agent status as an SSE event to all subscribers.
func (t *tailer) broadcastAgentStatus(projectPath string, status agentStatusJSON) {
	eventID := t.nextEventID()
	var message strings.Builder
	fmt.Fprintf(&message, "id: %d\n", eventID)

	// Wrap status in type envelope
	envelope := map[string]interface{}{
		"type":        "agent",
		"provisioned": status.Provisioned,
		"running":     status.Running,
	}
	if status.Pid != 0 {
		envelope["pid"] = status.Pid
	}
	if status.LastLog != "" {
		envelope["last_log"] = status.LastLog
	}

	line, err := json.Marshal(envelope)
	if err != nil {
		return
	}
	fmt.Fprintf(&message, "data: %s\n\n", string(line))

	t.broadcastToSubscribers(projectPath, message.String(), true)
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

	// Register subscriber (tickets only, no agent status)
	ch := make(chan string, 10) // Buffered for backpressure handling
	sub := t.subscribe(projectPath, ch, false)
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

// handleStatusSubscribe handles SSE subscriptions for both tickets and agent status.
func handleStatusSubscribe(w http.ResponseWriter, r *http.Request) {
	handleStatusSubscribeWithTailer(w, r, globalTailer)
}

// handleStatusSubscribeWithTailer handles SSE subscriptions with tickets and agent status.
func handleStatusSubscribeWithTailer(w http.ResponseWriter, r *http.Request, t *tailer) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract project from URL path or query parameter
	// Path format: /status/{project} where project is a #tag
	// Query format: /status?project=/absolute/path
	var projectPath string

	if q := r.URL.Query().Get("project"); q != "" {
		projectPath = q
	} else {
		path := strings.TrimPrefix(r.URL.Path, "/status/")
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

	// Get initial snapshot of tickets
	ticketsDir := projectPath + "/.ko/tickets"
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}
	SortByPriorityThenModified(tickets)

	// Get agent status
	agentStatus := getAgentStatus(ticketsDir)

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

	// First, send agent status with type discriminator
	agentEnvelope := map[string]interface{}{
		"type":        "agent",
		"provisioned": agentStatus.Provisioned,
		"running":     agentStatus.Running,
	}
	if agentStatus.Pid != 0 {
		agentEnvelope["pid"] = agentStatus.Pid
	}
	if agentStatus.LastLog != "" {
		agentEnvelope["last_log"] = agentStatus.LastLog
	}
	agentLine, jsonErr := json.Marshal(agentEnvelope)
	if jsonErr == nil {
		fmt.Fprintf(w, "data: %s\n", string(agentLine))
	}

	// Then send tickets with type discriminator
	for _, ticket := range tickets {
		j := ticketToJSON(ticket, ticketsDir)
		ticketEnvelope := map[string]interface{}{
			"type":   "ticket",
			"ticket": j,
		}
		line, jsonErr := json.Marshal(ticketEnvelope)
		if jsonErr != nil {
			continue
		}
		fmt.Fprintf(w, "data: %s\n", string(line))
	}
	fmt.Fprintf(w, "\n")
	flusher.Flush()

	// Register subscriber (includes both tickets and agent status)
	ch := make(chan string, 10) // Buffered for backpressure handling
	sub := t.subscribe(projectPath, ch, true)
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
