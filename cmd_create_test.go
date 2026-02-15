package main

import "testing"

func TestDerivePrefix(t *testing.T) {
	tests := []struct {
		dirName string
		want    string
	}{
		{"my-cool-project", "mcp"},
		{"fort-nix", "fn"},
		{"fort_nix", "fn"},
		{"exocortex", "exo"},
		{"knockout", "kno"},
		{"tk", "tk"},
		{"a", "ko"},         // too short, ultimate fallback
		{"my_cool_project", "mcp"},
		{"CamelCase", "cam"}, // single segment, lowercased
		{"A-B", "ab"},
		{"hello-world-app", "hwa"},
	}
	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			got := DerivePrefix(tt.dirName)
			if got != tt.want {
				t.Errorf("DerivePrefix(%q) = %q, want %q", tt.dirName, got, tt.want)
			}
		})
	}
}
