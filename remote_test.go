package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRemoteExec_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/ko" {
			t.Errorf("expected /ko, got %s", r.URL.Path)
		}

		// Verify request body
		var req struct {
			Argv []string `json:"argv"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Argv) != 1 || req.Argv[0] != "ls" {
			t.Errorf("expected argv=[ls], got %v", req.Argv)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ticket-1 open\nticket-2 closed\n")
	}))
	defer srv.Close()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := remoteExec(srv.URL, []string{"ls"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "ticket-1 open") {
		t.Errorf("expected output to contain ticket listing, got %q", string(out))
	}
}

func TestRemoteExec_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ticket not found\n"})
	}))
	defer srv.Close()

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := remoteExec(srv.URL, []string{"show", "nonexistent"})

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "ticket not found") {
		t.Errorf("expected stderr to contain error message, got %q", string(out))
	}
}

func TestRemoteExec_ConnectionFailure(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := remoteExec("http://localhost:1", []string{"ls"})

	w.Close()
	os.Stderr = old

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "server unreachable") {
		t.Errorf("expected 'server unreachable' error, got %q", string(out))
	}
}

func TestRemoteExec_TrailingSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ko" {
			t.Errorf("expected /ko, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Server URL with trailing slash should still work
	code := remoteExec(srv.URL+"/", []string{"ls"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestIsRemoteCommand(t *testing.T) {
	// Should proxy
	for _, cmd := range []string{"ls", "show", "add", "update", "start", "close", "open", "dep", "undep", "note", "bump", "status", "ready", "triage", "block", "snooze", "project", "stats", "search", "history"} {
		if !isRemoteCommand(cmd) {
			t.Errorf("expected %q to be a remote command", cmd)
		}
	}

	// Should NOT proxy
	for _, cmd := range []string{"serve", "agent", "help", "--help", "-h", "version", "--version", "-v", "import"} {
		if isRemoteCommand(cmd) {
			t.Errorf("expected %q to be a local-only command", cmd)
		}
	}
}

func TestGlobalConfig_Server(t *testing.T) {
	cfg, err := ParseGlobalConfig("server: https://ko.gisi.network")
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}
	if cfg.Server != "https://ko.gisi.network" {
		t.Errorf("expected server=https://ko.gisi.network, got %q", cfg.Server)
	}
}

func TestGlobalConfig_ServerWithComment(t *testing.T) {
	cfg, err := ParseGlobalConfig("server: https://ko.gisi.network # remote knockout")
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}
	if cfg.Server != "https://ko.gisi.network" {
		t.Errorf("expected server=https://ko.gisi.network, got %q", cfg.Server)
	}
}

func TestGlobalConfig_NoServer(t *testing.T) {
	cfg, err := ParseGlobalConfig("summarizer: ollama run qwen3:0.6b")
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}
	if cfg.Server != "" {
		t.Errorf("expected empty server, got %q", cfg.Server)
	}
}
