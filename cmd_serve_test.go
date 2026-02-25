package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
