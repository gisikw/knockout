package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePlanQuestions(t *testing.T) {
	tests := []struct {
		name      string
		questions []PlanQuestion
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty slice is valid",
			questions: []PlanQuestion{},
			wantErr:   false,
		},
		{
			name: "valid question with minimal fields",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
						{Label: "Option B", Value: "b"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid question with all fields",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Context:  "Some context",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a", Description: "Description A"},
						{Label: "Option B", Value: "b", Description: "Description B"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid questions",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "First question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
				{
					ID:       "q2",
					Question: "Second question?",
					Options: []QuestionOption{
						{Label: "Option X", Value: "x"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing id",
			questions: []PlanQuestion{
				{
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'id'",
		},
		{
			name: "missing question",
			questions: []PlanQuestion{
				{
					ID: "q1",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'question'",
		},
		{
			name: "missing options",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options:  []QuestionOption{},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'options'",
		},
		{
			name: "option missing label",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'label'",
		},
		{
			name: "option missing value",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'value'",
		},
		{
			name: "second question invalid",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Valid question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
				{
					ID: "q2",
					// Missing question field
					Options: []QuestionOption{
						{Label: "Option B", Value: "b"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'question'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlanQuestions(tt.questions)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlanQuestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePlanQuestions() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestCmdStatus(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantErr    bool
	}{
		{
			name: "set status to open",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "captured",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0001", "open"},
			wantStatus: "open",
		},
		{
			name: "set status to closed",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "in_progress",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0002", "closed"},
			wantStatus: "closed",
		},
		{
			name: "set status to blocked",
			ticket: &Ticket{
				ID:       "test-0003",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0003", "blocked"},
			wantStatus: "blocked",
		},
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "missing status",
			ticket: &Ticket{
				ID:       "test-0004",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:    []string{"test-0004"},
			wantErr: true,
		},
		{
			name: "invalid status",
			ticket: &Ticket{
				ID:       "test-0005",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:    []string{"test-0005", "invalid_status"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Suppress stderr
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdStatus(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdStatus() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress stdout
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdStatus(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdStatus() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}
		})
	}
}

func TestCmdStart(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantErr    bool
	}{
		{
			name: "start ticket from open",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0001"},
			wantStatus: "in_progress",
		},
		{
			name: "start ticket from captured",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "captured",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0002"},
			wantStatus: "in_progress",
		},
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Suppress stderr
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdStart(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdStart() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress stdout
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdStart(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdStart() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}
		})
	}
}

func TestCmdClose(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantErr    bool
	}{
		{
			name: "close ticket from in_progress",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "in_progress",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0001"},
			wantStatus: "closed",
		},
		{
			name: "close ticket from open",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0002"},
			wantStatus: "closed",
		},
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Suppress stderr
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdClose(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdClose() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress stdout
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdClose(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdClose() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}
		})
	}
}

func TestCmdOpen(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantErr    bool
	}{
		{
			name: "open ticket from captured",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "captured",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0001"},
			wantStatus: "open",
		},
		{
			name: "open ticket from blocked",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0002"},
			wantStatus: "open",
		},
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Suppress stderr
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdOpen(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdOpen() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress stdout
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdOpen(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdOpen() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}
		})
	}
}

func TestCmdSnooze(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantSnooze string
		wantErr    bool
	}{
		{
			name: "valid date sets snooze field",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0001", "2026-05-01"},
			wantSnooze: "2026-05-01",
		},
		{
			name:    "missing date arg returns error",
			args:    []string{"test-0001"},
			wantErr: true,
		},
		{
			name: "invalid date format returns error",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:    []string{"test-0002", "not-a-date"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdSnooze(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdSnooze() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdSnooze(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdSnooze() = %d, want 0", exitCode)
				return
			}

			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Snooze != tt.wantSnooze {
				t.Errorf("Snooze = %q, want %q", updated.Snooze, tt.wantSnooze)
			}
		})
	}
}

func TestCmdBlock(t *testing.T) {
	tests := []struct {
		name           string
		ticket         *Ticket
		args           []string
		wantStatus     string
		wantBodyText   string
		wantQuestions  int
		wantErr        bool
	}{
		{
			name: "block with reason",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:         []string{"test-0001", "waiting for API changes"},
			wantStatus:   "blocked",
			wantBodyText: "waiting for API changes",
		},
		{
			name: "block with questions",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args: []string{"test-0002", "--questions", `[{"id":"q1","question":"How to proceed?","options":[{"label":"Option A","value":"a"},{"label":"Option B","value":"b"}]}]`},
			wantStatus:    "blocked",
			wantQuestions: 1,
		},
		{
			name: "block without reason",
			ticket: &Ticket{
				ID:       "test-0003",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
			},
			args:       []string{"test-0003"},
			wantStatus: "blocked",
		},
		{
			name:    "missing ticket ID",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Suppress stderr
				oldStderr := os.Stderr
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()

				exitCode := cmdBlock(tt.args)
				if exitCode == 0 {
					t.Errorf("cmdBlock() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			tmpDir := t.TempDir()
			ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
			if err := os.MkdirAll(ticketsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := SaveTicket(ticketsDir, tt.ticket); err != nil {
				t.Fatal(err)
			}

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Suppress stdout
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = oldStdout }()

			exitCode := cmdBlock(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdBlock() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}

			if tt.wantBodyText != "" && !strings.Contains(updated.Body, tt.wantBodyText) {
				t.Errorf("Body does not contain %q\nBody: %s", tt.wantBodyText, updated.Body)
			}

			if tt.wantQuestions > 0 && len(updated.PlanQuestions) != tt.wantQuestions {
				t.Errorf("len(PlanQuestions) = %d, want %d", len(updated.PlanQuestions), tt.wantQuestions)
			}
		})
	}
}
