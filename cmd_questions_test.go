package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCmdQuestions(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		wantErr    bool
		wantOutput []PlanQuestion
	}{
		{
			name: "ticket with plan-questions",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "",
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "Should we keep backwards compatibility with pipeline.yml?",
						Context:  "INVARIANTS.md says...",
						Options: []QuestionOption{
							{Label: "Deprecation path (Recommended)", Value: "deprecate", Description: "Add warning, support both"},
							{Label: "Hard break", Value: "hard_break", Description: "Remove pipeline.yml support"},
						},
					},
					{
						ID:       "q2",
						Question: "Which test framework?",
						Options: []QuestionOption{
							{Label: "testscript", Value: "testscript"},
							{Label: "table tests", Value: "table"},
						},
					},
				},
			},
			wantErr: false,
			wantOutput: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Should we keep backwards compatibility with pipeline.yml?",
					Context:  "INVARIANTS.md says...",
					Options: []QuestionOption{
						{Label: "Deprecation path (Recommended)", Value: "deprecate", Description: "Add warning, support both"},
						{Label: "Hard break", Value: "hard_break", Description: "Remove pipeline.yml support"},
					},
				},
				{
					ID:       "q2",
					Question: "Which test framework?",
					Options: []QuestionOption{
						{Label: "testscript", Value: "testscript"},
						{Label: "table tests", Value: "table"},
					},
				},
			},
		},
		{
			name: "ticket with no plan-questions",
			ticket: &Ticket{
				ID:            "test-0002",
				Status:        "open",
				Deps:          []string{},
				Created:       "2026-01-01T00:00:00Z",
				Type:          "task",
				Priority:      2,
				Title:         "Test ticket",
				Body:          "",
				PlanQuestions: []PlanQuestion{},
			},
			wantErr:    false,
			wantOutput: []PlanQuestion{},
		},
		{
			name: "ticket with nil plan-questions",
			ticket: &Ticket{
				ID:            "test-0003",
				Status:        "open",
				Deps:          []string{},
				Created:       "2026-01-01T00:00:00Z",
				Type:          "task",
				Priority:      2,
				Title:         "Test ticket",
				Body:          "",
				PlanQuestions: nil,
			},
			wantErr:    false,
			wantOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Save the test ticket
			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			// Change to the temp directory so FindTicketsDir works
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

			// Run cmdQuestions
			args := []string{tt.ticket.ID}
			exitCode := cmdQuestions(args)

			// Close writer and read output
			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdQuestions() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdQuestions() = %d, want 0", exitCode)
				return
			}

			// Parse the JSON output
			var got []PlanQuestion
			if err := json.Unmarshal([]byte(output), &got); err != nil {
				t.Fatalf("failed to unmarshal output: %v\nOutput: %s", err, output)
			}

			// Compare the output
			if len(got) != len(tt.wantOutput) {
				t.Errorf("len(output) = %d, want %d", len(got), len(tt.wantOutput))
				return
			}

			for i := range got {
				if got[i].ID != tt.wantOutput[i].ID {
					t.Errorf("questions[%d].ID = %q, want %q", i, got[i].ID, tt.wantOutput[i].ID)
				}
				if got[i].Question != tt.wantOutput[i].Question {
					t.Errorf("questions[%d].Question = %q, want %q", i, got[i].Question, tt.wantOutput[i].Question)
				}
				if got[i].Context != tt.wantOutput[i].Context {
					t.Errorf("questions[%d].Context = %q, want %q", i, got[i].Context, tt.wantOutput[i].Context)
				}
				if len(got[i].Options) != len(tt.wantOutput[i].Options) {
					t.Errorf("len(questions[%d].Options) = %d, want %d", i, len(got[i].Options), len(tt.wantOutput[i].Options))
				}
			}
		})
	}
}

func TestCmdQuestionsErrors(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupFunc func(tmpDir string) error
		wantErr   bool
	}{
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "nonexistent ticket ID",
			args:    []string{"test-9999"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
				// Create a different ticket
				ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
				if err := os.MkdirAll(ticketsDir, 0755); err != nil {
					return err
				}
				ticket := &Ticket{
					ID:       "test-1234",
					Status:   "open",
					Deps:     []string{},
					Created:  "2026-01-01T00:00:00Z",
					Type:     "task",
					Priority: 2,
					Title:    "Other ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tmpDir := t.TempDir()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(tmpDir); err != nil {
					t.Fatal(err)
				}
			}

			// Change to the temp directory
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			os.Stdout, _ = os.Open(os.DevNull)
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			// Run cmdQuestions
			exitCode := cmdQuestions(tt.args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdQuestions() = %d, want non-zero exit code", exitCode)
				}
			} else {
				if exitCode != 0 {
					t.Errorf("cmdQuestions() = %d, want 0", exitCode)
				}
			}
		})
	}
}

func TestCmdQuestionsPartialID(t *testing.T) {
	// Create a temporary directory for this test
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test ticket
	ticket := &Ticket{
		ID:       "test-abcd",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-01-01T00:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Test ticket",
		Body:     "",
		PlanQuestions: []PlanQuestion{
			{
				ID:       "q1",
				Question: "Test question?",
				Options: []QuestionOption{
					{Label: "Yes", Value: "yes"},
					{Label: "No", Value: "no"},
				},
			},
		},
	}

	if err := SaveTicket(ticketsDir, ticket); err != nil {
		t.Fatal(err)
	}

	// Change to the temp directory
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

	// Run cmdQuestions with partial ID
	args := []string{"abcd"}
	exitCode := cmdQuestions(args)

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("cmdQuestions() = %d, want 0", exitCode)
		return
	}

	// Parse the JSON output
	var got []PlanQuestion
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("failed to unmarshal output: %v\nOutput: %s", err, output)
	}

	// Verify we got the question
	if len(got) != 1 {
		t.Errorf("len(output) = %d, want 1", len(got))
	}
	if len(got) > 0 && got[0].ID != "q1" {
		t.Errorf("question ID = %q, want %q", got[0].ID, "q1")
	}
}
