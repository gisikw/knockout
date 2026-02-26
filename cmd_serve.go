package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func cmdServe(args []string) int {
	// Parse flags
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	port := fs.String("port", "9876", "port to listen on")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Define whitelist of allowed subcommands
	whitelist := map[string]bool{
		"add":     true,
		"show":    true,
		"ls":      true,
		"ready":   true,
		"update":  true,
		"status":  true,
		"start":   true,
		"close":   true,
		"open":    true,
		"dep":     true,
		"undep":   true,
		"note":    true,
		"bump":    true,
		"agent":   true,
		"project": true,
	}

	// Start global event tailer
	globalTailer.start()

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe/", handleSubscribe)
	mux.HandleFunc("/status/", handleStatusSubscribe)
	mux.HandleFunc("/ko", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var req struct {
			Argv []string `json:"argv"`
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

		// Execute command
		cmd := exec.Command(os.Args[0], req.Argv...)
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
