package main

import "testing"

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
