package main

import (
	"os"
	"testing"
)

func TestReadonlyEnabled(t *testing.T) {
	orig := os.Getenv(readonlyEnvVar)
	defer os.Setenv(readonlyEnvVar, orig)

	cases := map[string]bool{
		"":      false,
		"0":     false,
		"false": false,
		"1":     true,
		"true":  true,
		"yes":   true,
	}
	for val, want := range cases {
		os.Setenv(readonlyEnvVar, val)
		if got := ReadonlyEnabled(); got != want {
			t.Errorf("ReadonlyEnabled with %q = %v, want %v", val, got, want)
		}
	}
}

func TestIsLegacyWrite(t *testing.T) {
	tests := []struct {
		cmd  string
		rest []string
		want bool
	}{
		{"add", []string{"a title"}, true},
		{"update", []string{"ko-1", "--status=open"}, true},
		{"close", []string{"ko-1"}, true},
		{"start", []string{"ko-1"}, true},
		{"open", []string{"ko-1"}, true},
		{"block", []string{"ko-1"}, true},
		{"snooze", []string{"ko-1", "2026-01-01"}, true},
		{"dep", []string{"ko-1", "ko-2"}, true},
		{"undep", []string{"ko-1", "ko-2"}, true},
		{"note", []string{"ko-1", "hi"}, true},
		{"bump", []string{"ko-1"}, true},
		// Reads and read sub-modes:
		{"show", []string{"ko-1"}, false},
		{"ls", nil, false},
		{"ready", nil, false},
		{"search", []string{"x"}, false},
		{"stats", nil, false},
		{"history", nil, false},
		{"dep", []string{"tree", "ko-1"}, false},    // dep tree is a read
		{"triage", []string{"ko-1"}, false},         // list mode (read)
		{"triage", nil, false},                      // list mode (read)
		{"triage", []string{"ko-1", "do it"}, true}, // set mode (write)
	}
	for _, tc := range tests {
		if got := isLegacyWrite(tc.cmd, tc.rest); got != tc.want {
			t.Errorf("isLegacyWrite(%q, %v) = %v, want %v", tc.cmd, tc.rest, got, tc.want)
		}
	}
}
