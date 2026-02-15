package main

import (
	"testing"
)

func TestParseRegistry(t *testing.T) {
	input := `default: exo
projects:
  exo: /home/dev/Projects/exocortex
  fort-nix: /home/dev/Projects/fort-nix
`
	reg, err := ParseRegistry(input)
	if err != nil {
		t.Fatalf("ParseRegistry: %v", err)
	}
	if reg.Default != "exo" {
		t.Errorf("Default = %q, want %q", reg.Default, "exo")
	}
	if len(reg.Projects) != 2 {
		t.Fatalf("len(Projects) = %d, want 2", len(reg.Projects))
	}
	if reg.Projects["exo"] != "/home/dev/Projects/exocortex" {
		t.Errorf("Projects[exo] = %q", reg.Projects["exo"])
	}
	if reg.Projects["fort-nix"] != "/home/dev/Projects/fort-nix" {
		t.Errorf("Projects[fort-nix] = %q", reg.Projects["fort-nix"])
	}
}

func TestParseRegistryNoDefault(t *testing.T) {
	input := `projects:
  exo: /tmp/exo
`
	reg, err := ParseRegistry(input)
	if err != nil {
		t.Fatalf("ParseRegistry: %v", err)
	}
	if reg.Default != "" {
		t.Errorf("Default = %q, want empty", reg.Default)
	}
	if reg.Projects["exo"] != "/tmp/exo" {
		t.Errorf("Projects[exo] = %q", reg.Projects["exo"])
	}
}

func TestFormatRegistryRoundTrip(t *testing.T) {
	reg := &Registry{
		Default: "exo",
		Projects: map[string]string{
			"exo":      "/tmp/exo",
			"fort-nix": "/tmp/fort-nix",
		},
	}
	output := FormatRegistry(reg)
	parsed, err := ParseRegistry(output)
	if err != nil {
		t.Fatalf("ParseRegistry round-trip: %v", err)
	}
	if parsed.Default != reg.Default {
		t.Errorf("Default = %q, want %q", parsed.Default, reg.Default)
	}
	if len(parsed.Projects) != len(reg.Projects) {
		t.Fatalf("len(Projects) = %d, want %d", len(parsed.Projects), len(reg.Projects))
	}
	for k, v := range reg.Projects {
		if parsed.Projects[k] != v {
			t.Errorf("Projects[%s] = %q, want %q", k, parsed.Projects[k], v)
		}
	}
}

func TestCleanTag(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"#fort-nix", "fort-nix"},
		{"fort-nix", "fort-nix"},
		{"#exo", "exo"},
		{"##double", "#double"},
	}
	for _, tt := range tests {
		got := CleanTag(tt.input)
		if got != tt.want {
			t.Errorf("CleanTag(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
