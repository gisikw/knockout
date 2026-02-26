package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdTriageBare(t *testing.T) {
	tests := []struct {
		name             string
		ticket           *Ticket
		wantBlockReason  string
		wantQuestionsLen int
	}{
		{
			name: "blocked with reason and questions",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "## Notes\n\n**2026-01-01 00:00:00 UTC:** ko: BLOCKED — missing requirements",
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
			},
			wantBlockReason:  "missing requirements",
			wantQuestionsLen: 1,
		},
		{
			name: "blocked without reason but with questions",
			ticket: &Ticket{
				ID:       "test-0002",
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
						Question: "Test question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			wantBlockReason:  "",
			wantQuestionsLen: 1,
		},
		{
			name: "open ticket with no questions",
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
			wantBlockReason:  "",
			wantQuestionsLen: 0,
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

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			args := []string{tt.ticket.ID}
			exitCode := cmdTriage(args)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("cmdTriage() = %d, want 0", exitCode)
				return
			}

			// Check for block reason in output
			if tt.wantBlockReason != "" {
				if !strings.Contains(output, tt.wantBlockReason) {
					t.Errorf("output does not contain block reason %q\nOutput: %s", tt.wantBlockReason, output)
				}
			}

			// Extract and parse the JSON part of the output
			jsonStart := strings.Index(output, "[")
			if jsonStart < 0 {
				t.Fatal("no JSON array found in output")
			}
			jsonOutput := output[jsonStart:]

			var questions []PlanQuestion
			if err := json.Unmarshal([]byte(jsonOutput), &questions); err != nil {
				t.Fatalf("failed to unmarshal questions: %v\nJSON: %s", err, jsonOutput)
			}

			if len(questions) != tt.wantQuestionsLen {
				t.Errorf("len(questions) = %d, want %d", len(questions), tt.wantQuestionsLen)
			}
		})
	}
}

func TestCmdTriageJSON(t *testing.T) {
	tests := []struct {
		name              string
		ticket            *Ticket
		wantBlockReason   string
		wantQuestionsLen  int
	}{
		{
			name: "blocked with reason and questions",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "## Notes\n\n**2026-01-01 00:00:00 UTC:** ko: BLOCKED — missing requirements",
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
			},
			wantBlockReason:  "missing requirements",
			wantQuestionsLen: 1,
		},
		{
			name: "blocked without reason but with questions",
			ticket: &Ticket{
				ID:       "test-0002",
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
						Question: "Test question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			wantBlockReason:  "",
			wantQuestionsLen: 1,
		},
		{
			name: "open ticket with no questions",
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
			wantBlockReason:  "",
			wantQuestionsLen: 0,
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

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			args := []string{tt.ticket.ID, "--json"}
			exitCode := cmdTriage(args)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("cmdTriage() = %d, want 0", exitCode)
				return
			}

			// Parse JSON output
			var state triageStateJSON
			if err := json.Unmarshal([]byte(output), &state); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v\nOutput: %s", err, output)
			}

			if state.BlockReason != tt.wantBlockReason {
				t.Errorf("BlockReason = %q, want %q", state.BlockReason, tt.wantBlockReason)
			}

			if len(state.Questions) != tt.wantQuestionsLen {
				t.Errorf("len(Questions) = %d, want %d", len(state.Questions), tt.wantQuestionsLen)
			}
		})
	}
}

func TestCmdTriageBlock(t *testing.T) {
	tests := []struct {
		name       string
		ticket     *Ticket
		args       []string
		wantStatus string
		wantNote   string
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
				Title:    "Test ticket",
				Body:     "",
			},
			args:       []string{"test-0001", "--block", "waiting for approval"},
			wantStatus: "blocked",
			wantNote:   "ko: BLOCKED — waiting for approval",
		},
		{
			name: "block without reason",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "",
			},
			args:       []string{"test-0002", "--block"},
			wantStatus: "blocked",
			wantNote:   "",
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

			exitCode := cmdTriage(tt.args)

			if exitCode != 0 {
				t.Errorf("cmdTriage() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", updated.Status, tt.wantStatus)
			}

			if tt.wantNote != "" && !strings.Contains(updated.Body, tt.wantNote) {
				t.Errorf("Body does not contain note %q\nBody: %s", tt.wantNote, updated.Body)
			}
		})
	}
}

func TestCmdTriageQuestions(t *testing.T) {
	tests := []struct {
		name         string
		ticket       *Ticket
		questionsJSON string
		wantErr      bool
		wantStatus   string
		wantQCount   int
	}{
		{
			name: "add questions - auto-blocks",
			ticket: &Ticket{
				ID:       "test-0001",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "",
			},
			questionsJSON: `[{"id":"q1","question":"Test?","options":[{"label":"Yes","value":"yes"},{"label":"No","value":"no"}]}]`,
			wantErr:       false,
			wantStatus:    "blocked",
			wantQCount:    1,
		},
		{
			name: "invalid JSON",
			ticket: &Ticket{
				ID:       "test-0002",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "",
			},
			questionsJSON: `{invalid}`,
			wantErr:       true,
		},
		{
			name: "invalid questions - missing id",
			ticket: &Ticket{
				ID:       "test-0003",
				Status:   "open",
				Deps:     []string{},
				Created:  "2026-01-01T00:00:00Z",
				Type:     "task",
				Priority: 2,
				Title:    "Test ticket",
				Body:     "",
			},
			questionsJSON: `[{"question":"Test?","options":[{"label":"Yes","value":"yes"}]}]`,
			wantErr:       true,
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

			args := []string{tt.ticket.ID, "--questions", tt.questionsJSON}
			exitCode := cmdTriage(args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdTriage() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdTriage() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", updated.Status, tt.wantStatus)
			}

			if len(updated.PlanQuestions) != tt.wantQCount {
				t.Errorf("len(PlanQuestions) = %d, want %d", len(updated.PlanQuestions), tt.wantQCount)
			}
		})
	}
}

func TestCmdTriageAnswers(t *testing.T) {
	tests := []struct {
		name               string
		ticket             *Ticket
		answersJSON        string
		wantErr            bool
		wantStatus         string
		wantQuestionsCount int
		wantNotesContain   []string
	}{
		{
			name: "partial answer - questions remain",
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
						Question: "First question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes", Description: "Go with yes"},
							{Label: "No", Value: "no"},
						},
					},
					{
						ID:       "q2",
						Question: "Second question?",
						Options: []QuestionOption{
							{Label: "A", Value: "a"},
							{Label: "B", Value: "b"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"yes"}`,
			wantErr:            false,
			wantStatus:         "blocked",
			wantQuestionsCount: 1,
			wantNotesContain: []string{
				"Question: First question?\nAnswer: Yes\nGo with yes",
			},
		},
		{
			name: "full answer - all questions resolved, auto-unblock",
			ticket: &Ticket{
				ID:       "test-0002",
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
						Question: "First question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
							{Label: "No", Value: "no"},
						},
					},
					{
						ID:       "q2",
						Question: "Second question?",
						Options: []QuestionOption{
							{Label: "A", Value: "a"},
							{Label: "B", Value: "b"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"yes","q2":"a"}`,
			wantErr:            false,
			wantStatus:         "open",
			wantQuestionsCount: 0,
			wantNotesContain: []string{
				"Question: First question?\nAnswer: Yes",
				"Question: Second question?\nAnswer: A",
			},
		},
		{
			name: "invalid JSON",
			ticket: &Ticket{
				ID:       "test-0003",
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
						Question: "Test?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			answersJSON: `{invalid}`,
			wantErr:     true,
		},
		{
			name: "question ID not found",
			ticket: &Ticket{
				ID:       "test-0004",
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
						Question: "Test?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			answersJSON: `{"q99":"yes"}`,
			wantErr:     true,
		},
		{
			name: "no questions to answer",
			ticket: &Ticket{
				ID:            "test-0005",
				Status:        "open",
				Deps:          []string{},
				Created:       "2026-01-01T00:00:00Z",
				Type:          "task",
				Priority:      2,
				Title:         "Test ticket",
				Body:          "",
				PlanQuestions: []PlanQuestion{},
			},
			answersJSON: `{"q1":"yes"}`,
			wantErr:     true,
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

			args := []string{tt.ticket.ID, "--answers", tt.answersJSON}
			exitCode := cmdTriage(args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdTriage() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdTriage() = %d, want 0", exitCode)
				return
			}

			// Load updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			if updated.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", updated.Status, tt.wantStatus)
			}

			if len(updated.PlanQuestions) != tt.wantQuestionsCount {
				t.Errorf("len(PlanQuestions) = %d, want %d", len(updated.PlanQuestions), tt.wantQuestionsCount)
			}

			for _, noteText := range tt.wantNotesContain {
				if !strings.Contains(updated.Body, noteText) {
					t.Errorf("Body does not contain expected note: %q\nBody: %s", noteText, updated.Body)
				}
			}

			if len(tt.wantNotesContain) > 0 && !strings.Contains(updated.Body, "## Notes") {
				t.Error("Body does not contain ## Notes section")
			}
		})
	}
}

func TestCmdTriageErrors(t *testing.T) {
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
			name:    "nonexistent ticket",
			args:    []string{"test-9999"},
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
					Title:    "Other ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
		{
			name:    "missing --questions argument",
			args:    []string{"test-1234", "--questions"},
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
					Title:    "Test ticket",
					Body:     "",
				}
				return SaveTicket(ticketsDir, ticket)
			},
		},
		{
			name:    "missing --answers argument",
			args:    []string{"test-1234", "--answers"},
			wantErr: true,
			setupFunc: func(tmpDir string) error {
				ticketsDir := filepath.Join(tmpDir, ".ko", "tickets")
				if err := os.MkdirAll(ticketsDir, 0755); err != nil {
					return err
				}
				ticket := &Ticket{
					ID:       "test-1234",
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

			// Suppress output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			os.Stdout, _ = os.Open(os.DevNull)
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			exitCode := cmdTriage(tt.args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdTriage() = %d, want non-zero exit code", exitCode)
				}
			} else {
				if exitCode != 0 {
					t.Errorf("cmdTriage() = %d, want 0", exitCode)
				}
			}
		})
	}
}
