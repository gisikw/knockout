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
