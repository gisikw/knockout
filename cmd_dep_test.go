package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCmdDepTreeJSON(t *testing.T) {
	tests := []struct {
		name           string
		tickets        []*Ticket
		rootID         string
		wantID         string
		wantStatus     string
		wantTitle      string
		wantDepsCount  int
		checkFirstDep  bool
		firstDepID     string
		firstDepStatus string
	}{
		{
			name: "single node",
			tickets: []*Ticket{
				{
					ID:       "test-0001",
					Status:   "open",
					Deps:     []string{},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Single task",
					Body:     "",
				},
			},
			rootID:        "test-0001",
			wantID:        "test-0001",
			wantStatus:    "open",
			wantTitle:     "Single task",
			wantDepsCount: 0,
		},
		{
			name: "linear chain",
			tickets: []*Ticket{
				{
					ID:       "test-0001",
					Status:   "open",
					Deps:     []string{"test-0002"},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Root task",
					Body:     "",
				},
				{
					ID:       "test-0002",
					Status:   "closed",
					Deps:     []string{},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Dependency",
					Body:     "",
				},
			},
			rootID:         "test-0001",
			wantID:         "test-0001",
			wantStatus:     "open",
			wantTitle:      "Root task",
			wantDepsCount:  1,
			checkFirstDep:  true,
			firstDepID:     "test-0002",
			firstDepStatus: "closed",
		},
		{
			name: "branching tree",
			tickets: []*Ticket{
				{
					ID:       "test-0001",
					Status:   "open",
					Deps:     []string{"test-0002", "test-0003"},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Root task",
					Body:     "",
				},
				{
					ID:       "test-0002",
					Status:   "open",
					Deps:     []string{},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "First dep",
					Body:     "",
				},
				{
					ID:       "test-0003",
					Status:   "closed",
					Deps:     []string{},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Second dep",
					Body:     "",
				},
			},
			rootID:        "test-0001",
			wantID:        "test-0001",
			wantStatus:    "open",
			wantTitle:     "Root task",
			wantDepsCount: 2,
		},
		{
			name: "cycle detection",
			tickets: []*Ticket{
				{
					ID:       "test-0001",
					Status:   "open",
					Deps:     []string{"test-0002"},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Task A",
					Body:     "",
				},
				{
					ID:       "test-0002",
					Status:   "open",
					Deps:     []string{"test-0001"},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Task B",
					Body:     "",
				},
			},
			rootID:         "test-0001",
			wantID:         "test-0001",
			wantStatus:     "open",
			wantTitle:      "Task A",
			wantDepsCount:  1,
			checkFirstDep:  true,
			firstDepID:     "test-0002",
			firstDepStatus: "open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			for _, ticket := range tt.tickets {
				if err := SaveTicket(ticketsDir, ticket); err != nil {
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

			args := []string{"tree", tt.rootID, "--json"}
			exitCode := cmdDep(args)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("cmdDepTree() = %d, want 0", exitCode)
				return
			}

			// Parse JSON output
			var tree depTreeJSON
			if err := json.Unmarshal([]byte(output), &tree); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v\nOutput: %s", err, output)
			}

			if tree.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", tree.ID, tt.wantID)
			}

			if tree.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", tree.Status, tt.wantStatus)
			}

			if tree.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", tree.Title, tt.wantTitle)
			}

			if len(tree.Deps) != tt.wantDepsCount {
				t.Errorf("len(Deps) = %d, want %d", len(tree.Deps), tt.wantDepsCount)
			}

			if tt.checkFirstDep && len(tree.Deps) > 0 {
				firstDep := tree.Deps[0]
				if firstDep.ID != tt.firstDepID {
					t.Errorf("FirstDep.ID = %q, want %q", firstDep.ID, tt.firstDepID)
				}
				if firstDep.Status != tt.firstDepStatus {
					t.Errorf("FirstDep.Status = %q, want %q", firstDep.Status, tt.firstDepStatus)
				}
			}
		})
	}
}
