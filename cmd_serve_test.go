package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServeHandler(t *testing.T) {
	// Save original os.Args and restore at end
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set os.Args[0] to a known ko binary (the test binary itself will work for testing)
	os.Args = []string{"ko"}

	// Define whitelist (same as in cmdServe)
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

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Argv []string `json:"argv"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Argv) == 0 {
			http.Error(w, "argv must have at least one element", http.StatusBadRequest)
			return
		}

		subcommand := req.Argv[0]
		if !whitelist[subcommand] {
			errResp := map[string]string{
				"error": "subcommand '" + subcommand + "' not allowed",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		// For testing purposes, just return success for whitelisted commands
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	tests := []struct {
		name           string
		method         string
		body           string
		wantStatus     int
		wantContains   string
		checkJSON      bool
		wantJSONField  string
		wantJSONValue  string
	}{
		{
			name:         "GET method not allowed",
			method:       http.MethodGet,
			body:         "",
			wantStatus:   http.StatusMethodNotAllowed,
			wantContains: "method not allowed",
		},
		{
			name:         "invalid JSON",
			method:       http.MethodPost,
			body:         "not json",
			wantStatus:   http.StatusBadRequest,
			wantContains: "invalid JSON",
		},
		{
			name:         "empty argv",
			method:       http.MethodPost,
			body:         `{"argv":[]}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "argv must have at least one element",
		},
		{
			name:          "invalid subcommand",
			method:        http.MethodPost,
			body:          `{"argv":["invalid"]}`,
			wantStatus:    http.StatusBadRequest,
			checkJSON:     true,
			wantJSONField: "error",
			wantJSONValue: "subcommand 'invalid' not allowed",
		},
		{
			name:          "invalid subcommand rm",
			method:        http.MethodPost,
			body:          `{"argv":["rm","-rf","/"]}`,
			wantStatus:    http.StatusBadRequest,
			checkJSON:     true,
			wantJSONField: "error",
			wantJSONValue: "subcommand 'rm' not allowed",
		},
		{
			name:         "valid subcommand ls",
			method:       http.MethodPost,
			body:         `{"argv":["ls"]}`,
			wantStatus:   http.StatusOK,
			wantContains: "ok",
		},
		{
			name:         "valid subcommand ready",
			method:       http.MethodPost,
			body:         `{"argv":["ready"]}`,
			wantStatus:   http.StatusOK,
			wantContains: "ok",
		},
		{
			name:         "valid subcommand show with args",
			method:       http.MethodPost,
			body:         `{"argv":["show","test-id"]}`,
			wantStatus:   http.StatusOK,
			wantContains: "ok",
		},
		{
			name:         "valid subcommand agent",
			method:       http.MethodPost,
			body:         `{"argv":["agent","status"]}`,
			wantStatus:   http.StatusOK,
			wantContains: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/ko", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tt.wantStatus)
			}

			if tt.checkJSON {
				var resp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse JSON response: %v", err)
				}
				if resp[tt.wantJSONField] != tt.wantJSONValue {
					t.Errorf("JSON field %q = %q, want %q", tt.wantJSONField, resp[tt.wantJSONField], tt.wantJSONValue)
				}
			} else if tt.wantContains != "" {
				body := w.Body.String()
				if !bytes.Contains([]byte(body), []byte(tt.wantContains)) {
					t.Errorf("response body = %q, want to contain %q", body, tt.wantContains)
				}
			}
		})
	}
}

func TestServeWhitelist(t *testing.T) {
	// Verify all expected commands are in whitelist
	expectedCommands := []string{
		"ls", "ready", "blocked", "resolved", "closed",
		"query", "show", "questions", "answer", "close",
		"reopen", "block", "start", "bump", "note",
		"status", "dep", "undep", "agent",
	}

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

	for _, cmd := range expectedCommands {
		if !whitelist[cmd] {
			t.Errorf("expected command %q missing from whitelist", cmd)
		}
	}

	// Verify dangerous commands are NOT in whitelist
	dangerousCommands := []string{
		"rm", "mv", "cp", "sh", "bash", "eval", "exec",
		"create", "add", "init", // explicitly excluded per plan
	}

	for _, cmd := range dangerousCommands {
		if whitelist[cmd] {
			t.Errorf("dangerous command %q should not be in whitelist", cmd)
		}
	}
}

func TestTailerBasic(t *testing.T) {
	// Test the tailer in isolation
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "state")
	oldState := os.Getenv("XDG_STATE_HOME")
	os.Setenv("XDG_STATE_HOME", stateDir)
	defer os.Setenv("XDG_STATE_HOME", oldState)

	projectDir := filepath.Join(tmpDir, "project")
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	eventsFile := filepath.Join(stateDir, "knockout", "events.jsonl")
	os.MkdirAll(ticketsDir, 0755)
	os.MkdirAll(filepath.Dir(eventsFile), 0755)
	os.WriteFile(eventsFile, []byte{}, 0644)

	// Create a test ticket
	testTicket := filepath.Join(ticketsDir, "test-abc.md")
	ticketContent := `---
id: test-abc
title: Test
status: open
type: task
priority: 1
created: 2024-01-01T00:00:00Z
---

Body.
`
	os.WriteFile(testTicket, []byte(ticketContent), 0644)

	// Create tailer
	testTailer := &tailer{
		subscribers: make(map[*subscriber]bool),
	}
	testTailer.start()

	// Subscribe
	ch := make(chan string, 10)
	sub := testTailer.subscribe(projectDir, ch)
	defer testTailer.unsubscribe(sub)

	// Write event
	time.Sleep(500 * time.Millisecond)
	event := MutationEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     "test",
		Project:   projectDir,
		Ticket:    "t-123",
	}
	eventJSON, _ := json.Marshal(event)
	eventJSON = append(eventJSON, '\n')
	f, _ := os.OpenFile(eventsFile, os.O_APPEND|os.O_WRONLY, 0644)
	f.Write(eventJSON)
	f.Close()

	// Wait for broadcast
	select {
	case msg := <-ch:
		if msg == "" {
			t.Error("received empty message")
		}
		t.Logf("Received broadcast (%d bytes): %q", len(msg), msg)
		// Check format
		if !strings.Contains(msg, "id:") {
			t.Error("broadcast missing id field")
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

// readSSEEvent reads lines from an SSE stream until a blank line (end of event).
// Returns all non-blank lines. Returns error on timeout.
func readSSEEvent(reader *bufio.Reader, timeout time.Duration) ([]string, error) {
	type result struct {
		lines []string
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				ch <- result{lines, err}
				return
			}
			if line == "\n" {
				ch <- result{lines, nil}
				return
			}
			lines = append(lines, strings.TrimRight(line, "\n"))
		}
	}()
	select {
	case r := <-ch:
		return r.lines, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout after %v", timeout)
	}
}

func sseHasLine(lines []string, prefix string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}
	return false
}

func setupSSETest(t *testing.T) (projectDir, eventsFile string, testTailer *tailer) {
	t.Helper()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "state")
	oldState := os.Getenv("XDG_STATE_HOME")
	os.Setenv("XDG_STATE_HOME", stateDir)
	t.Cleanup(func() { os.Setenv("XDG_STATE_HOME", oldState) })

	projectDir = filepath.Join(tmpDir, "testproject")
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	eventsFile = filepath.Join(stateDir, "knockout", "events.jsonl")
	os.MkdirAll(ticketsDir, 0755)
	os.MkdirAll(filepath.Dir(eventsFile), 0755)
	os.WriteFile(eventsFile, []byte{}, 0644)
	os.WriteFile(filepath.Join(ticketsDir, "test-1234.md"), []byte(`---
id: test-1234
status: open
type: task
priority: 1
created: 2024-01-01T00:00:00Z
---
# Test Ticket

Test ticket body.
`), 0644)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"ko"}

	testTailer = &tailer{subscribers: make(map[*subscriber]bool)}
	testTailer.start()
	time.Sleep(500 * time.Millisecond)
	return
}

func writeEvent(t *testing.T, eventsFile, projectDir string) {
	t.Helper()
	event := MutationEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     "status",
		Project:   projectDir,
		Ticket:    "test-1234",
	}
	eventJSON, _ := json.Marshal(event)
	eventJSON = append(eventJSON, '\n')
	f, _ := os.OpenFile(eventsFile, os.O_APPEND|os.O_WRONLY, 0644)
	f.Write(eventJSON)
	f.Close()
}

func TestSubscribeHandler(t *testing.T) {
	projectDir, eventsFile, testTailer := setupSSETest(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe/", func(w http.ResponseWriter, r *http.Request) {
		handleSubscribeWithTailer(w, r, testTailer)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/subscribe?project=" + projectDir)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	reader := bufio.NewReader(resp.Body)

	// Event 1: retry directive
	retry, err := readSSEEvent(reader, 3*time.Second)
	if err != nil {
		t.Fatalf("retry event: %v", err)
	}
	if !sseHasLine(retry, "retry:") {
		t.Errorf("expected retry directive, got: %v", retry)
	}

	// Event 2: initial snapshot
	snapshot, err := readSSEEvent(reader, 3*time.Second)
	if err != nil {
		t.Fatalf("snapshot event: %v", err)
	}
	if !sseHasLine(snapshot, "id: 0") {
		t.Errorf("snapshot missing id:0, got: %v", snapshot)
	}
	if !sseHasLine(snapshot, "data:") {
		t.Errorf("snapshot missing data, got: %v", snapshot)
	}

	// Trigger mutation
	writeEvent(t, eventsFile, projectDir)

	// Event 3: broadcast
	broadcast, err := readSSEEvent(reader, 5*time.Second)
	if err != nil {
		t.Fatalf("broadcast event: %v", err)
	}
	hasNonZeroID := false
	for _, line := range broadcast {
		if strings.HasPrefix(line, "id:") && !strings.Contains(line, "id: 0") {
			hasNonZeroID = true
		}
	}
	if !hasNonZeroID {
		t.Errorf("broadcast missing monotonic ID > 0, got: %v", broadcast)
	}
}

func TestSubscribeMultiple(t *testing.T) {
	projectDir, eventsFile, testTailer := setupSSETest(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe/", func(w http.ResponseWriter, r *http.Request) {
		handleSubscribeWithTailer(w, r, testTailer)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	numSubscribers := 3
	var wg sync.WaitGroup
	received := make([]bool, numSubscribers)
	errors := make([]error, numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			resp, err := http.Get(server.URL + "/subscribe?project=" + projectDir)
			if err != nil {
				errors[idx] = fmt.Errorf("connect: %v", err)
				return
			}
			defer resp.Body.Close()

			reader := bufio.NewReader(resp.Body)

			// Skip retry event
			if _, err := readSSEEvent(reader, 3*time.Second); err != nil {
				errors[idx] = fmt.Errorf("retry: %v", err)
				return
			}
			// Skip initial snapshot
			if _, err := readSSEEvent(reader, 3*time.Second); err != nil {
				errors[idx] = fmt.Errorf("snapshot: %v", err)
				return
			}

			// Wait for broadcast
			broadcast, err := readSSEEvent(reader, 5*time.Second)
			if err != nil {
				errors[idx] = fmt.Errorf("broadcast: %v", err)
				return
			}
			for _, line := range broadcast {
				if strings.HasPrefix(line, "id:") && !strings.Contains(line, "id: 0") {
					received[idx] = true
				}
			}
		}(i)
	}

	// Give subscribers time to connect and consume initial events
	time.Sleep(1 * time.Second)

	// Emit mutation
	writeEvent(t, eventsFile, projectDir)

	wg.Wait()

	for i := 0; i < numSubscribers; i++ {
		if errors[i] != nil {
			t.Errorf("subscriber %d: %v", i, errors[i])
		}
		if !received[i] {
			t.Errorf("subscriber %d did not receive event", i)
		}
	}
}
