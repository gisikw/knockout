package main

import (
	"strings"
	"testing"
	"time"
)

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

func TestIsSnoozed(t *testing.T) {
	// Fixed "now" for determinism: 2026-03-15T12:00:00Z
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		snooze string
		want   bool
	}{
		{
			name:   "empty snooze is not snoozed",
			snooze: "",
			want:   false,
		},
		{
			name:   "past date is not snoozed",
			snooze: "2020-01-01",
			want:   false,
		},
		{
			name:   "future date is snoozed",
			snooze: "2099-01-01",
			want:   true,
		},
		{
			name:   "today at midnight is not snoozed (valid as of midnight)",
			snooze: "2026-03-15",
			want:   false,
		},
		{
			name:   "invalid string is not snoozed",
			snooze: "not-a-date",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSnoozed(tt.snooze, now)
			if got != tt.want {
				t.Errorf("IsSnoozed(%q, %v) = %v, want %v", tt.snooze, now, got, tt.want)
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

func TestSnoozeRoundTrip(t *testing.T) {
	original := &Ticket{
		ID:       "test-1234",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-05-01T00:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Snoozed Ticket",
		Snooze:   "2026-05-01",
	}

	formatted := FormatTicket(original)

	if !strings.Contains(formatted, "snooze: 2026-05-01") {
		t.Errorf("FormatTicket output does not contain 'snooze: 2026-05-01'\noutput:\n%s", formatted)
	}

	parsed, err := ParseTicket(formatted)
	if err != nil {
		t.Fatalf("ParseTicket() error = %v", err)
	}

	if parsed.Snooze != "2026-05-01" {
		t.Errorf("Snooze = %q, want %q", parsed.Snooze, "2026-05-01")
	}
}

func TestTriageRoundTrip(t *testing.T) {
	original := &Ticket{
		ID:       "test-1234",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-05-01T00:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Triage Ticket",
		Triage:   "unblock this ticket",
	}

	formatted := FormatTicket(original)

	if !strings.Contains(formatted, "triage: unblock this ticket") {
		t.Errorf("FormatTicket output does not contain 'triage: unblock this ticket'\noutput:\n%s", formatted)
	}

	parsed, err := ParseTicket(formatted)
	if err != nil {
		t.Fatalf("ParseTicket() error = %v", err)
	}

	if parsed.Triage != "unblock this ticket" {
		t.Errorf("Triage = %q, want %q", parsed.Triage, "unblock this ticket")
	}
}

func TestParseTicketWithTriage(t *testing.T) {
	input := `---
id: test-1234
status: open
deps: []
created: 2026-05-01T00:00:00Z
type: task
priority: 2
triage: break this apart
---
# Triage Ticket
`
	parsed, err := ParseTicket(input)
	if err != nil {
		t.Fatalf("ParseTicket() error = %v", err)
	}
	if parsed.Triage != "break this apart" {
		t.Errorf("Triage = %q, want %q", parsed.Triage, "break this apart")
	}
}

func TestParseTicketWithSnooze(t *testing.T) {
	input := `---
id: test-1234
status: open
deps: []
created: 2026-05-01T00:00:00Z
type: task
priority: 2
snooze: 2026-05-01
---
# Snoozed Ticket
`
	parsed, err := ParseTicket(input)
	if err != nil {
		t.Fatalf("ParseTicket() error = %v", err)
	}
	if parsed.Snooze != "2026-05-01" {
		t.Errorf("Snooze = %q, want %q", parsed.Snooze, "2026-05-01")
	}
}

func TestSortByPriorityThenModified(t *testing.T) {
	// Create a base time for consistent testing
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		tickets  []*Ticket
		expected []string // expected order of ticket IDs
	}{
		{
			name: "in_progress sorts before open within same priority",
			tickets: []*Ticket{
				{ID: "open-1", Status: "open", Priority: 1, ModTime: baseTime},
				{ID: "in_progress-1", Status: "in_progress", Priority: 1, ModTime: baseTime},
			},
			expected: []string{"in_progress-1", "open-1"},
		},
		{
			name: "priority takes precedence over status",
			tickets: []*Ticket{
				{ID: "open-p2", Status: "open", Priority: 2, ModTime: baseTime},
				{ID: "in_progress-p1", Status: "in_progress", Priority: 1, ModTime: baseTime},
			},
			expected: []string{"in_progress-p1", "open-p2"},
		},
		{
			name: "modtime breaks ties within same priority and status",
			tickets: []*Ticket{
				{ID: "open-older", Status: "open", Priority: 1, ModTime: baseTime.Add(-1 * time.Hour)},
				{ID: "open-newer", Status: "open", Priority: 1, ModTime: baseTime},
			},
			expected: []string{"open-newer", "open-older"},
		},
		{
			name: "full ordering: priority, then status, then modtime",
			tickets: []*Ticket{
				{ID: "open-p2-newer", Status: "open", Priority: 2, ModTime: baseTime},
				{ID: "closed-p1-newest", Status: "closed", Priority: 1, ModTime: baseTime.Add(1 * time.Hour)},
				{ID: "in_progress-p1-older", Status: "in_progress", Priority: 1, ModTime: baseTime.Add(-1 * time.Hour)},
				{ID: "open-p1-newest", Status: "open", Priority: 1, ModTime: baseTime.Add(2 * time.Hour)},
				{ID: "resolved-p1-newer", Status: "resolved", Priority: 1, ModTime: baseTime},
				{ID: "in_progress-p1-newer", Status: "in_progress", Priority: 1, ModTime: baseTime},
			},
			expected: []string{
				"in_progress-p1-newer",   // p1, in_progress (0), newest of in_progress
				"in_progress-p1-older",   // p1, in_progress (0), older
				"open-p1-newest",         // p1, open (1), newest
				"resolved-p1-newer",      // p1, resolved (2)
				"closed-p1-newest",       // p1, closed (3)
				"open-p2-newer",          // p2, open (1)
			},
		},
		{
			name: "blocked status sorts after open",
			tickets: []*Ticket{
				{ID: "blocked-1", Status: "blocked", Priority: 1, ModTime: baseTime},
				{ID: "open-1", Status: "open", Priority: 1, ModTime: baseTime},
				{ID: "in_progress-1", Status: "in_progress", Priority: 1, ModTime: baseTime},
			},
			expected: []string{"in_progress-1", "open-1", "blocked-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test data
			tickets := make([]*Ticket, len(tt.tickets))
			copy(tickets, tt.tickets)

			// Sort the tickets
			SortByPriorityThenModified(tickets)

			// Verify the order
			if len(tickets) != len(tt.expected) {
				t.Fatalf("got %d tickets, want %d", len(tickets), len(tt.expected))
			}

			for i, ticket := range tickets {
				if ticket.ID != tt.expected[i] {
					t.Errorf("position %d: got ticket %q, want %q", i, ticket.ID, tt.expected[i])
				}
			}
		})
	}
}
