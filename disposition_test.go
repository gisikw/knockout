package main

import "testing"

func TestExtractLastFencedJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantOK  bool
	}{
		{
			name:   "single fenced block",
			input:  "thinking...\n```json\n{\"disposition\": \"continue\"}\n```\n",
			want:   `{"disposition": "continue"}`,
			wantOK: true,
		},
		{
			name:   "plain fence without json tag",
			input:  "thinking...\n```\n{\"disposition\": \"fail\"}\n```\n",
			want:   `{"disposition": "fail"}`,
			wantOK: true,
		},
		{
			name: "multiple blocks returns last",
			input: "first:\n```json\n{\"disposition\": \"continue\"}\n```\n" +
				"second:\n```json\n{\"disposition\": \"route\", \"workflow\": \"feature\"}\n```\n",
			want:   `{"disposition": "route", "workflow": "feature"}`,
			wantOK: true,
		},
		{
			name:   "no fenced block",
			input:  "just plain text\nno json here",
			want:   "",
			wantOK: false,
		},
		{
			name:   "unclosed fence",
			input:  "```json\n{\"disposition\": \"continue\"}\n",
			want:   "",
			wantOK: false,
		},
		{
			name:   "empty fenced block",
			input:  "```json\n```\n",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ExtractLastFencedJSON(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDisposition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantType string
		wantErr bool
	}{
		{
			name:     "continue",
			input:    `{"disposition": "continue"}`,
			wantType: "continue",
		},
		{
			name:     "fail with reason",
			input:    `{"disposition": "fail", "reason": "bad input"}`,
			wantType: "fail",
		},
		{
			name:     "blocked",
			input:    `{"disposition": "blocked", "block_on": "ko-123", "reason": "needs work"}`,
			wantType: "blocked",
		},
		{
			name:     "route",
			input:    `{"disposition": "route", "workflow": "feature"}`,
			wantType: "route",
		},
		{
			name:     "decompose",
			input:    `{"disposition": "decompose", "subtasks": ["a", "b"]}`,
			wantType: "decompose",
		},
		{
			name:    "invalid json",
			input:   `{not json}`,
			wantErr: true,
		},
		{
			name:    "missing disposition field",
			input:   `{"reason": "no type"}`,
			wantErr: true,
		},
		{
			name:    "unknown disposition type",
			input:   `{"disposition": "explode"}`,
			wantErr: true,
		},
		{
			name:    "route without workflow",
			input:   `{"disposition": "route"}`,
			wantErr: true,
		},
		{
			name:    "decompose without subtasks",
			input:   `{"disposition": "decompose"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ParseDisposition(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Type != tt.wantType {
				t.Errorf("type = %q, want %q", d.Type, tt.wantType)
			}
		})
	}
}
