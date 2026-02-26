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

func TestCmdAgentReportJSON(t *testing.T) {
	tests := []struct {
		name      string
		setupLog  bool
		logLines  []string
		wantEmpty bool
		wantData  *agentReportJSON
	}{
		{
			name:      "no log file",
			setupLog:  false,
			wantEmpty: true,
		},
		{
			name:      "no JSONL lines",
			setupLog:  true,
			logLines:  []string{"loop: building ko-abc", "loop: building ko-def"},
			wantEmpty: true,
		},
		{
			name:     "single JSONL line",
			setupLog: true,
			logLines: []string{
				"loop: building ko-abc",
				`{"ts":"2026-01-01T12:00:00Z","tickets_processed":5,"succeeded":3,"failed":1,"blocked":0,"decomposed":1,"stop_reason":"empty","runtime_seconds":120.5}`,
			},
			wantEmpty: false,
			wantData: &agentReportJSON{
				Timestamp:        "2026-01-01T12:00:00Z",
				TicketsProcessed: 5,
				Succeeded:        3,
				Failed:           1,
				Blocked:          0,
				Decomposed:       1,
				StopReason:       "empty",
				RuntimeSeconds:   120.5,
			},
		},
		{
			name:     "multiple JSONL lines - last one selected",
			setupLog: true,
			logLines: []string{
				"loop: first run",
				`{"ts":"2026-01-01T10:00:00Z","tickets_processed":2,"succeeded":2,"failed":0,"blocked":0,"decomposed":0,"stop_reason":"empty","runtime_seconds":60.0}`,
				"loop: second run",
				`{"ts":"2026-01-01T12:00:00Z","tickets_processed":5,"succeeded":3,"failed":1,"blocked":0,"decomposed":1,"stop_reason":"max_tickets","runtime_seconds":120.5}`,
			},
			wantEmpty: false,
			wantData: &agentReportJSON{
				Timestamp:        "2026-01-01T12:00:00Z",
				TicketsProcessed: 5,
				Succeeded:        3,
				Failed:           1,
				Blocked:          0,
				Decomposed:       1,
				StopReason:       "max_tickets",
				RuntimeSeconds:   120.5,
			},
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

			// Setup log file if requested
			if tt.setupLog {
				logPath := filepath.Join(koDir, "agent.log")
				content := ""
				for _, line := range tt.logLines {
					content += line + "\n"
				}
				if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
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
			exitCode := cmdAgentReport(args)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("cmdAgentReport() = %d, want 0", exitCode)
				return
			}

			// Parse JSON output
			var report agentReportJSON
			if err := json.Unmarshal([]byte(output), &report); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v\nOutput: %s", err, output)
			}

			if tt.wantEmpty {
				// Expect empty object
				if report.Timestamp != "" || report.TicketsProcessed != 0 {
					t.Errorf("expected empty report, got %+v", report)
				}
			} else {
				// Verify fields match expected data
				if report.Timestamp != tt.wantData.Timestamp {
					t.Errorf("Timestamp = %s, want %s", report.Timestamp, tt.wantData.Timestamp)
				}
				if report.TicketsProcessed != tt.wantData.TicketsProcessed {
					t.Errorf("TicketsProcessed = %d, want %d", report.TicketsProcessed, tt.wantData.TicketsProcessed)
				}
				if report.Succeeded != tt.wantData.Succeeded {
					t.Errorf("Succeeded = %d, want %d", report.Succeeded, tt.wantData.Succeeded)
				}
				if report.Failed != tt.wantData.Failed {
					t.Errorf("Failed = %d, want %d", report.Failed, tt.wantData.Failed)
				}
				if report.Blocked != tt.wantData.Blocked {
					t.Errorf("Blocked = %d, want %d", report.Blocked, tt.wantData.Blocked)
				}
				if report.Decomposed != tt.wantData.Decomposed {
					t.Errorf("Decomposed = %d, want %d", report.Decomposed, tt.wantData.Decomposed)
				}
				if report.StopReason != tt.wantData.StopReason {
					t.Errorf("StopReason = %s, want %s", report.StopReason, tt.wantData.StopReason)
				}
				if report.RuntimeSeconds != tt.wantData.RuntimeSeconds {
					t.Errorf("RuntimeSeconds = %f, want %f", report.RuntimeSeconds, tt.wantData.RuntimeSeconds)
				}
			}
		})
	}
}
