package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentSpawnHandler(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup project with config
	projectDir := filepath.Join(tmpDir, "testproject")
	koDir := filepath.Join(projectDir, ".ko")
	ticketsDir := filepath.Join(koDir, "tickets")
	os.MkdirAll(ticketsDir, 0755)
	os.WriteFile(filepath.Join(koDir, "config.yaml"), []byte(`workflows:
  main:
    - node: test
      type: decision
      prompt: test
`), 0644)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantField  string
	}{
		{
			name:       "missing project",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantField:  "error",
		},
		{
			name:       "invalid project",
			body:       `{"project": "#nonexistent"}`,
			wantStatus: http.StatusBadRequest,
			wantField:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/agent/spawn", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handleAgentSpawn(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantField != "" {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("invalid JSON response: %v", err)
				}
				if _, ok := resp[tt.wantField]; !ok {
					t.Errorf("response missing field %q: %v", tt.wantField, resp)
				}
			}
		})
	}

	// Test GET method not allowed
	t.Run("GET not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/agent/spawn", nil)
		w := httptest.NewRecorder()
		handleAgentSpawn(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestAgentKillHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "missing project",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid project",
			body:       `{"project": "#nonexistent"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/agent/kill", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handleAgentKill(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}

	// Test not running returns 404 for valid project
	t.Run("not running", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "proj")
		os.MkdirAll(projectDir, 0755)

		body := fmt.Sprintf(`{"project": "%s"}`, projectDir)
		req := httptest.NewRequest(http.MethodPost, "/agent/kill", bytes.NewBufferString(body))
		w := httptest.NewRecorder()

		handleAgentKill(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestAgentStatusHandler(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "proj")
	os.MkdirAll(projectDir, 0755)

	// Test not running
	t.Run("not running", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/agent/status?project="+projectDir, nil)
		w := httptest.NewRecorder()

		handleAgentStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if resp["running"] != false {
			t.Errorf("running = %v, want false", resp["running"])
		}
	})

	// Test POST not allowed
	t.Run("POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/agent/status?project="+projectDir, nil)
		w := httptest.NewRecorder()
		handleAgentStatus(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	// Test missing project
	t.Run("missing project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/agent/status", nil)
		w := httptest.NewRecorder()
		handleAgentStatus(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
