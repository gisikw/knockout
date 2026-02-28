package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdUpdateBasicFields(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantTitle  string
		wantType   string
		wantPrio   int
		wantAssign string
	}{
		{
			name: "update title",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Old Title",
				Body:     "",
			},
			args:      []string{"test-0001", "--title", "New Title"},
			wantTitle: "New Title",
			wantType:  "task",
			wantPrio:  2,
		},
		{
			name: "update type and priority",
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
			args:      []string{"test-0002", "-t", "bug", "-p", "0"},
			wantTitle: "Test Ticket",
			wantType:  "bug",
			wantPrio:  0,
		},
		{
			name: "update assignee",
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
			args:       []string{"test-0003", "-a", "alice"},
			wantTitle:  "Test Ticket",
			wantType:   "task",
			wantPrio:   2,
			wantAssign: "alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			exitCode := cmdUpdate(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", updated.Title, tt.wantTitle)
			}
			if updated.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", updated.Type, tt.wantType)
			}
			if updated.Priority != tt.wantPrio {
				t.Errorf("Priority = %d, want %d", updated.Priority, tt.wantPrio)
			}
			if tt.wantAssign != "" && updated.Assignee != tt.wantAssign {
				t.Errorf("Assignee = %q, want %q", updated.Assignee, tt.wantAssign)
			}
		})
	}
}

func TestCmdUpdateTags(t *testing.T) {
	tests := []struct {
		name        string
		ticket      *Ticket
		args        []string
		wantTags    []string
		description string
	}{
		{
			name: "replace tags",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
				Tags:     []string{"old", "tags"},
			},
			args:        []string{"test-0001", "--tags", "new,tags"},
			wantTags:    []string{"new", "tags"},
			description: "tags should replace, not append",
		},
		{
			name: "set tags on ticket with no tags",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
				Tags:     []string{},
			},
			args:        []string{"test-0002", "--tags", "foo,bar,baz"},
			wantTags:    []string{"foo", "bar", "baz"},
			description: "tags should be set correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			exitCode := cmdUpdate(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if len(updated.Tags) != len(tt.wantTags) {
				t.Errorf("len(Tags) = %d, want %d", len(updated.Tags), len(tt.wantTags))
			}

			for i, want := range tt.wantTags {
				if i >= len(updated.Tags) || updated.Tags[i] != want {
					t.Errorf("Tags[%d] = %q, want %q", i, updated.Tags[i], want)
				}
			}
		})
	}
}

func TestCmdUpdateStatus(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantErr    bool
	}{
		{
			name: "update status to closed",
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
			args:       []string{"test-0001", "--status", "closed"},
			wantStatus: "closed",
			wantErr:    false,
		},
		{
			name: "invalid status",
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
			args:    []string{"test-0002", "--status", "invalid_status"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Suppress stdout/stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			os.Stdout, _ = os.Open(os.DevNull)
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			exitCode := cmdUpdate(tt.args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdUpdate() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
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

func TestCmdUpdateAutoUnblock(t *testing.T) {
	tests := []struct {
		name               string
		ticket             *Ticket
		answersJSON        string
		wantStatus         string
		wantQuestionsCount int
	}{
		{
			name: "answer all questions - auto-unblock",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "First question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
							{Label: "No", Value: "no"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"yes"}`,
			wantStatus:         "open",
			wantQuestionsCount: 0,
		},
		{
			name: "answer partial questions - remain blocked",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
				Body:     "",
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "First question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
					{
						ID:       "q2",
						Question: "Second question?",
						Options: []QuestionOption{
							{Label: "A", Value: "a"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"yes"}`,
			wantStatus:         "blocked",
			wantQuestionsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			args := []string{tt.ticket.ID, "--answers", tt.answersJSON}
			exitCode := cmdUpdate(args)

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
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

			if len(updated.PlanQuestions) != tt.wantQuestionsCount {
				t.Errorf("len(PlanQuestions) = %d, want %d", len(updated.PlanQuestions), tt.wantQuestionsCount)
			}
		})
	}
}

func TestCmdUpdateErrors(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupFunc func(tmpDir string) error
		wantErr   bool
	}{
		{
			name:    "missing ticket ID",
			args:    []string{"--title", "New Title"},
			wantErr: true,
		},
		{
			name:    "nonexistent ticket",
			args:    []string{"test-9999", "--title", "New Title"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
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
					Title:    "Test Ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
		{
			name:    "no fields specified",
			args:    []string{"test-1234"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
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
					Title:    "Test Ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
		{
			name:    "invalid questions JSON",
			args:    []string{"test-1234", "--questions", "{invalid}"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
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
					Title:    "Test Ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
		{
			name:    "invalid answers JSON",
			args:    []string{"test-1234", "--answers", "{invalid}"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
				ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
				if err := os.MkdirAll(ticketsDir, 0755); err != nil {
					return err
				}
				ticket := &Ticket{
					ID:     "test-1234",
					Status: "blocked",
					Deps:   []string{},
					Created: "2026-01-01T00:00:00Z",
					Type:   "task",
					Priority: 2,
					Title:  "Test Ticket",
					Body:   "",
					PlanQuestions: []PlanQuestion{
						{
							ID:       "q1",
							Question: "Test?",
							Options: []QuestionOption{
								{Label: "Yes", Value: "yes"},
							},
						},
					},
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(tmpDir); err != nil {
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

			// Suppress stdout/stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			os.Stdout, _ = os.Open(os.DevNull)
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			exitCode := cmdUpdate(tt.args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdUpdate() = %d, want non-zero exit code", exitCode)
				}
			} else {
				if exitCode != 0 {
					t.Errorf("cmdUpdate() = %d, want 0", exitCode)
				}
			}
		})
	}
}

func TestCmdUpdateSnooze(t *testing.T) {
	tests := []struct {
		name        string
		ticket      *Ticket
		args        []string
		wantSnooze  string
		wantErr     bool
	}{
		{
			name: "set valid snooze date",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
			},
			args:       []string{"test-0001", "--snooze", "2026-05-01"},
			wantSnooze: "2026-05-01",
			wantErr:    false,
		},
		{
			name: "invalid snooze date rejected",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test Ticket",
			},
			args:    []string{"test-0002", "--snooze", "not-a-date"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			oldStderr := os.Stderr
			os.Stdout, _ = os.Open(os.DevNull)
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			exitCode := cmdUpdate(tt.args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdUpdate() = 0, want non-zero for invalid snooze")
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
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

func TestCmdUpdateDescriptionAndDesign(t *testing.T) {
	tests := []struct {
		name         string
		ticket       *Ticket
		args         []string
		wantBodyText string
	}{
		{
			name: "add description",
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
			args:         []string{"test-0001", "-d", "New description"},
			wantBodyText: "New description",
		},
		{
			name: "add design notes",
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
			args:         []string{"test-0002", "--design", "Design notes here"},
			wantBodyText: "## Design\n\nDesign notes here",
		},
		{
			name: "add acceptance criteria",
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
			args:         []string{"test-0003", "--acceptance", "Must pass all tests"},
			wantBodyText: "## Acceptance Criteria\n\nMust pass all tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			exitCode := cmdUpdate(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdUpdate() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(updated.Body, tt.wantBodyText) {
				t.Errorf("Body does not contain %q\nBody: %s", tt.wantBodyText, updated.Body)
			}
		})
	}
}
