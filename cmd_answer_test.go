package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdAnswer(t *testing.T) {
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
			name: "partial answer - some questions remain",
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
						Question: "Tabs or spaces?",
						Options: []QuestionOption{
							{Label: "Tabs", Value: "tabs"},
							{Label: "Spaces, 2-wide", Value: "spaces", Description: "Use 2-space indentation"},
						},
					},
					{
						ID:       "q2",
						Question: "Fix manually or with script?",
						Options: []QuestionOption{
							{Label: "Manual fix", Value: "manual", Description: "I will fix manually"},
							{Label: "Script", Value: "script"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"spaces"}`,
			wantErr:            false,
			wantStatus:         "blocked",
			wantQuestionsCount: 1,
			wantNotesContain: []string{
				"Question: Tabs or spaces?\nAnswer: Spaces, 2-wide\nUse 2-space indentation",
			},
		},
		{
			name: "full answer - all questions resolved",
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
						Question: "Tabs or spaces?",
						Options: []QuestionOption{
							{Label: "Tabs", Value: "tabs"},
							{Label: "Spaces, 2-wide", Value: "spaces", Description: "Use 2-space indentation"},
						},
					},
					{
						ID:       "q2",
						Question: "Fix manually or with script?",
						Options: []QuestionOption{
							{Label: "Manual fix", Value: "manual", Description: "I will fix manually"},
							{Label: "Script", Value: "script"},
						},
					},
				},
			},
			answersJSON:        `{"q1":"spaces","q2":"manual"}`,
			wantErr:            false,
			wantStatus:         "open",
			wantQuestionsCount: 0,
			wantNotesContain: []string{
				"Question: Tabs or spaces?\nAnswer: Spaces, 2-wide\nUse 2-space indentation",
				"Question: Fix manually or with script?\nAnswer: Manual fix\nI will fix manually",
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
						Question: "Test question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			answersJSON: `{invalid json}`,
			wantErr:     true,
		},
		{
			name: "nonexistent question ID",
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
						Question: "Test question?",
						Options: []QuestionOption{
							{Label: "Yes", Value: "yes"},
						},
					},
				},
			},
			answersJSON: `{"q99":"answer"}`,
			wantErr:     true,
		},
		{
			name: "ticket with no plan-questions",
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
			answersJSON: `{"q1":"answer"}`,
			wantErr:     true,
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

			// Run cmdAnswer
			args := []string{tt.ticket.ID, tt.answersJSON}
			exitCode := cmdAnswer(args)

			if tt.wantErr {
				if exitCode == 0 {
					t.Errorf("cmdAnswer() = %d, want non-zero exit code", exitCode)
				}
				return
			}

			if exitCode != 0 {
				t.Errorf("cmdAnswer() = %d, want 0", exitCode)
				return
			}

			// Load the updated ticket
			updated, err := LoadTicket(ticketsDir, tt.ticket.ID)
			if err != nil {
				t.Fatal(err)
			}

			// Check status
			if updated.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", updated.Status, tt.wantStatus)
			}

			// Check remaining questions count
			if len(updated.PlanQuestions) != tt.wantQuestionsCount {
				t.Errorf("len(PlanQuestions) = %d, want %d", len(updated.PlanQuestions), tt.wantQuestionsCount)
			}

			// Check that notes were added with correct format
			for _, noteText := range tt.wantNotesContain {
				if !strings.Contains(updated.Body, noteText) {
					t.Errorf("Body does not contain expected note text: %q", noteText)
				}
			}

			// Verify notes section exists
			if len(tt.wantNotesContain) > 0 && !strings.Contains(updated.Body, "## Notes") {
				t.Error("Body does not contain ## Notes section")
			}
		})
	}
}
