package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCmdAgentStatusJSON(t *testing.T) {
	tests := []struct {
		name            string
		setupPipeline   bool
		setupPid        bool
		pidAlive        bool
		wantProvisioned bool
		wantRunning     bool
	}{
		{
			name:            "not provisioned",
			setupPipeline:   false,
			setupPid:        false,
			pidAlive:        false,
			wantProvisioned: false,
			wantRunning:     false,
		},
		{
			name:            "provisioned but not running",
			setupPipeline:   true,
			setupPid:        false,
			pidAlive:        false,
			wantProvisioned: true,
			wantRunning:     false,
		},
		{
			name:            "provisioned and running",
			setupPipeline:   true,
			setupPid:        true,
			pidAlive:        true,
			wantProvisioned: true,
			wantRunning:     true,
		},
		{
			name:            "stale pid file",
			setupPipeline:   true,
			setupPid:        true,
			pidAlive:        false,
			wantProvisioned: true,
			wantRunning:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			koDir := filepath.Join(tmpDir, ".ko")
			ticketsDir := filepath.Join(koDir, "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Setup pipeline config if requested
			if tt.setupPipeline {
				configPath := filepath.Join(koDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte("workflows:\n  main:\n    - action: test\n      command: echo test\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Setup PID file if requested
			var pid int
			if tt.setupPid {
				if tt.pidAlive {
					// Use our own PID (guaranteed to be alive)
					pid = os.Getpid()
				} else {
					// Use a PID that definitely doesn't exist
					pid = 999999
				}
				pidPath := filepath.Join(koDir, "agent.pid")
				if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
					t.Fatal(err)
				}
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			args := []string{"--json"}
			exitCode := cmdAgentStatus(args)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("cmdAgentStatus() = %d, want 0", exitCode)
				return
			}

			// Parse JSON output
			var status agentStatusJSON
			if err := json.Unmarshal([]byte(output), &status); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v\nOutput: %s", err, output)
			}

			if status.Provisioned != tt.wantProvisioned {
				t.Errorf("Provisioned = %v, want %v", status.Provisioned, tt.wantProvisioned)
			}

			if status.Running != tt.wantRunning {
				t.Errorf("Running = %v, want %v", status.Running, tt.wantRunning)
			}

			if tt.wantRunning && status.Pid != pid {
				t.Errorf("Pid = %d, want %d", status.Pid, pid)
			}
		})
	}
}
