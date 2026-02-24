package main

import "testing"

func TestIsReady(t *testing.T) {
	tests := []struct {
		name            string
		status          string
		allDepsResolved bool
		want            bool
	}{
		{
			name:            "open with deps resolved",
			status:          "open",
			allDepsResolved: true,
			want:            true,
		},
		{
			name:            "open with deps unresolved",
			status:          "open",
			allDepsResolved: false,
			want:            false,
		},
		{
			name:            "in_progress with deps resolved",
			status:          "in_progress",
			allDepsResolved: true,
			want:            true,
		},
		{
			name:            "in_progress with deps unresolved",
			status:          "in_progress",
			allDepsResolved: false,
			want:            false,
		},
		{
			name:            "resolved is never ready",
			status:          "resolved",
			allDepsResolved: true,
			want:            false,
		},
		{
			name:            "closed is never ready",
			status:          "closed",
			allDepsResolved: true,
			want:            false,
		},
		{
			name:            "blocked is never ready",
			status:          "blocked",
			allDepsResolved: true,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsReady(tt.status, tt.allDepsResolved)
			if got != tt.want {
				t.Errorf("IsReady(%q, %v) = %v, want %v", tt.status, tt.allDepsResolved, got, tt.want)
			}
		})
	}
}

func TestExtractBlockReason(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "FAIL note with reason",
			body: "## Notes\n\n**2026-02-24 07:40:00 UTC:** ko: FAIL at node 'actionable' — Plan contains an open question\n",
			want: "Plan contains an open question",
		},
		{
			name: "BLOCKED note with reason",
			body: "## Notes\n\n**2026-02-24 08:00:00 UTC:** ko: BLOCKED at node 'verify' — Missing test coverage\n",
			want: "Missing test coverage",
		},
		{
			name: "multiple notes returns most recent",
			body: "## Notes\n\n**2026-02-24 07:00:00 UTC:** ko: FAIL at node 'plan' — First reason\n\n**2026-02-24 08:00:00 UTC:** ko: BLOCKED at node 'verify' — Second reason\n",
			want: "Second reason",
		},
		{
			name: "FAIL without reason",
			body: "## Notes\n\n**2026-02-24 07:40:00 UTC:** ko: FAIL at node 'actionable'\n",
			want: "",
		},
		{
			name: "no notes section",
			body: "Just some ticket body text",
			want: "",
		},
		{
			name: "notes section without FAIL/BLOCKED",
			body: "## Notes\n\n**2026-02-24 07:40:00 UTC:** Just a regular note\n",
			want: "",
		},
		{
			name: "empty body",
			body: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &Ticket{Body: tt.body}
			got := ExtractBlockReason(ticket)
			if got != tt.want {
				t.Errorf("ExtractBlockReason() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlanQuestions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    *Ticket
	}{
		{
			name: "empty questions array",
			input: `---
id: test-1234
status: blocked
deps: []
created: 2026-02-24T10:00:00Z
type: task
priority: 2
plan-questions:
---
# Test Ticket
Body content`,
			wantErr: false,
			want: &Ticket{
				ID:            "test-1234",
				Status:        "blocked",
				Deps:          []string{},
				Created:       "2026-02-24T10:00:00Z",
				Type:          "task",
				Priority:      2,
				PlanQuestions: []PlanQuestion{},
				Title:         "Test Ticket",
				Body:          "Body content",
			},
		},
		{
			name: "single question with two options",
			input: `---
id: test-1234
status: blocked
deps: []
created: 2026-02-24T10:00:00Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Tabs or spaces?"
    options:
      - label: "Spaces"
        value: spaces
      - label: "Tabs"
        value: tabs
---
# Test Ticket
Body content`,
			wantErr: false,
			want: &Ticket{
				ID:       "test-1234",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-02-24T10:00:00Z",
				Type:     "task",
				Priority: 2,
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "Tabs or spaces?",
						Options: []QuestionOption{
							{Label: "Spaces", Value: "spaces"},
							{Label: "Tabs", Value: "tabs"},
						},
					},
				},
				Title: "Test Ticket",
				Body:  "Body content",
			},
		},
		{
			name: "question with context and descriptions",
			input: `---
id: test-1234
status: blocked
deps: []
created: 2026-02-24T10:00:00Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Which library?"
    context: "INVARIANTS.md says no external deps"
    options:
      - label: "Standard library"
        value: stdlib
        description: "Matches invariant"
      - label: "External library"
        value: external
        description: "Violates invariant"
---
# Test Ticket
Body content`,
			wantErr: false,
			want: &Ticket{
				ID:       "test-1234",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-02-24T10:00:00Z",
				Type:     "task",
				Priority: 2,
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "Which library?",
						Context:  "INVARIANTS.md says no external deps",
						Options: []QuestionOption{
							{
								Label:       "Standard library",
								Value:       "stdlib",
								Description: "Matches invariant",
							},
							{
								Label:       "External library",
								Value:       "external",
								Description: "Violates invariant",
							},
						},
					},
				},
				Title: "Test Ticket",
				Body:  "Body content",
			},
		},
		{
			name: "multiple questions",
			input: `---
id: test-1234
status: blocked
deps: []
created: 2026-02-24T10:00:00Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "First question?"
    options:
      - label: "Option A"
        value: a
      - label: "Option B"
        value: b
  - id: q2
    question: "Second question?"
    options:
      - label: "Option X"
        value: x
      - label: "Option Y"
        value: y
---
# Test Ticket
Body content`,
			wantErr: false,
			want: &Ticket{
				ID:       "test-1234",
				Status:   "blocked",
				Deps:     []string{},
				Created:  "2026-02-24T10:00:00Z",
				Type:     "task",
				Priority: 2,
				PlanQuestions: []PlanQuestion{
					{
						ID:       "q1",
						Question: "First question?",
						Options: []QuestionOption{
							{Label: "Option A", Value: "a"},
							{Label: "Option B", Value: "b"},
						},
					},
					{
						ID:       "q2",
						Question: "Second question?",
						Options: []QuestionOption{
							{Label: "Option X", Value: "x"},
							{Label: "Option Y", Value: "y"},
						},
					},
				},
				Title: "Test Ticket",
				Body:  "Body content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTicket(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTicket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Compare fields
			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.Status != tt.want.Status {
				t.Errorf("Status = %q, want %q", got.Status, tt.want.Status)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.Body != tt.want.Body {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}

			// Compare PlanQuestions
			if len(got.PlanQuestions) != len(tt.want.PlanQuestions) {
				t.Errorf("PlanQuestions length = %d, want %d", len(got.PlanQuestions), len(tt.want.PlanQuestions))
				return
			}
			for i, q := range got.PlanQuestions {
				wantQ := tt.want.PlanQuestions[i]
				if q.ID != wantQ.ID {
					t.Errorf("PlanQuestions[%d].ID = %q, want %q", i, q.ID, wantQ.ID)
				}
				if q.Question != wantQ.Question {
					t.Errorf("PlanQuestions[%d].Question = %q, want %q", i, q.Question, wantQ.Question)
				}
				if q.Context != wantQ.Context {
					t.Errorf("PlanQuestions[%d].Context = %q, want %q", i, q.Context, wantQ.Context)
				}
				if len(q.Options) != len(wantQ.Options) {
					t.Errorf("PlanQuestions[%d].Options length = %d, want %d", i, len(q.Options), len(wantQ.Options))
					continue
				}
				for j, opt := range q.Options {
					wantOpt := wantQ.Options[j]
					if opt.Label != wantOpt.Label {
						t.Errorf("PlanQuestions[%d].Options[%d].Label = %q, want %q", i, j, opt.Label, wantOpt.Label)
					}
					if opt.Value != wantOpt.Value {
						t.Errorf("PlanQuestions[%d].Options[%d].Value = %q, want %q", i, j, opt.Value, wantOpt.Value)
					}
					if opt.Description != wantOpt.Description {
						t.Errorf("PlanQuestions[%d].Options[%d].Description = %q, want %q", i, j, opt.Description, wantOpt.Description)
					}
				}
			}
		})
	}
}

func TestPlanQuestionsRoundTrip(t *testing.T) {
	original := &Ticket{
		ID:       "test-1234",
		Status:   "blocked",
		Deps:     []string{},
		Created:  "2026-02-24T10:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Test Ticket",
		Body:     "Body content",
		PlanQuestions: []PlanQuestion{
			{
				ID:       "q1",
				Question: "Tabs or spaces?",
				Context:  "INVARIANTS.md says spaces",
				Options: []QuestionOption{
					{
						Label:       "Spaces, 2-wide (Recommended)",
						Value:       "spaces_2",
						Description: "Matches existing codebase convention",
					},
					{
						Label:       "Tabs",
						Value:       "tabs",
						Description: "Let the editor decide width",
					},
				},
			},
		},
	}

	// Format the ticket
	formatted := FormatTicket(original)

	// Parse it back
	parsed, err := ParseTicket(formatted)
	if err != nil {
		t.Fatalf("ParseTicket() error = %v", err)
	}

	// Compare
	if parsed.ID != original.ID {
		t.Errorf("Round trip: ID = %q, want %q", parsed.ID, original.ID)
	}
	if parsed.Status != original.Status {
		t.Errorf("Round trip: Status = %q, want %q", parsed.Status, original.Status)
	}
	if parsed.Title != original.Title {
		t.Errorf("Round trip: Title = %q, want %q", parsed.Title, original.Title)
	}
	if parsed.Body != original.Body {
		t.Errorf("Round trip: Body = %q, want %q", parsed.Body, original.Body)
	}

	if len(parsed.PlanQuestions) != len(original.PlanQuestions) {
		t.Fatalf("Round trip: PlanQuestions length = %d, want %d", len(parsed.PlanQuestions), len(original.PlanQuestions))
	}

	for i, q := range parsed.PlanQuestions {
		origQ := original.PlanQuestions[i]
		if q.ID != origQ.ID {
			t.Errorf("Round trip: PlanQuestions[%d].ID = %q, want %q", i, q.ID, origQ.ID)
		}
		if q.Question != origQ.Question {
			t.Errorf("Round trip: PlanQuestions[%d].Question = %q, want %q", i, q.Question, origQ.Question)
		}
		if q.Context != origQ.Context {
			t.Errorf("Round trip: PlanQuestions[%d].Context = %q, want %q", i, q.Context, origQ.Context)
		}
		if len(q.Options) != len(origQ.Options) {
			t.Errorf("Round trip: PlanQuestions[%d].Options length = %d, want %d", i, len(q.Options), len(origQ.Options))
			continue
		}
		for j, opt := range q.Options {
			origOpt := origQ.Options[j]
			if opt.Label != origOpt.Label {
				t.Errorf("Round trip: Options[%d].Label = %q, want %q", j, opt.Label, origOpt.Label)
			}
			if opt.Value != origOpt.Value {
				t.Errorf("Round trip: Options[%d].Value = %q, want %q", j, opt.Value, origOpt.Value)
			}
			if opt.Description != origOpt.Description {
				t.Errorf("Round trip: Options[%d].Description = %q, want %q", j, opt.Description, origOpt.Description)
			}
		}
	}
}
