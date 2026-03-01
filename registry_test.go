package main

import (
	"strings"
	"testing"
)

func TestParseRegistry(t *testing.T) {
	input := `projects:
  exo:
    path: /home/dev/Projects/exocortex
    default: true
  fort-nix:
    path: /home/dev/Projects/fort-nix
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
  exo:
    path: /tmp/exo
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

func TestParseRegistryNewFormat(t *testing.T) {
	input := `projects:
  exo:
    path: /tmp/exo
    prefix: exo
    default: true
  fort-nix:
    path: /tmp/fn
    prefix: fn
`
	reg, err := ParseRegistry(input)
	if err != nil {
		t.Fatalf("ParseRegistry: %v", err)
	}
	if reg.Default != "exo" {
		t.Errorf("Default = %q, want %q", reg.Default, "exo")
	}
	if reg.Projects["exo"] != "/tmp/exo" {
		t.Errorf("Projects[exo] = %q", reg.Projects["exo"])
	}
	if reg.Projects["fort-nix"] != "/tmp/fn" {
		t.Errorf("Projects[fort-nix] = %q", reg.Projects["fort-nix"])
	}
	if reg.Prefixes["exo"] != "exo" {
		t.Errorf("Prefixes[exo] = %q, want %q", reg.Prefixes["exo"], "exo")
	}
	if reg.Prefixes["fort-nix"] != "fn" {
		t.Errorf("Prefixes[fort-nix] = %q, want %q", reg.Prefixes["fort-nix"], "fn")
	}
}

func TestFormatRegistryRoundTrip(t *testing.T) {
	reg := &Registry{
		Default: "exo",
		Projects: map[string]string{
			"exo":      "/tmp/exo",
			"fort-nix": "/tmp/fort-nix",
			"secret":   "/tmp/secret",
		},
		Prefixes: map[string]string{
			"exo":      "exo",
			"fort-nix": "fn",
		},
		Hidden: map[string]bool{
			"secret": true,
		},
	}
	output := FormatRegistry(reg)
	if strings.Contains(output, "prefixes:") {
		t.Error("FormatRegistry output contains 'prefixes:' section (should use nested format)")
	}
	if strings.HasPrefix(output, "default:") {
		t.Error("FormatRegistry output starts with top-level 'default:' key")
	}
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
	if len(parsed.Prefixes) != len(reg.Prefixes) {
		t.Fatalf("len(Prefixes) = %d, want %d", len(parsed.Prefixes), len(reg.Prefixes))
	}
	for k, v := range reg.Prefixes {
		if parsed.Prefixes[k] != v {
			t.Errorf("Prefixes[%s] = %q, want %q", k, parsed.Prefixes[k], v)
		}
	}
	for k, v := range reg.Hidden {
		if parsed.Hidden[k] != v {
			t.Errorf("Hidden[%s] = %v, want %v", k, parsed.Hidden[k], v)
		}
	}
}

func TestParseRegistryBackwardCompatible(t *testing.T) {
	// Old flat format should still parse (backward compat preserved)
	input := `default: exo
projects:
  exo: /tmp/exo
`
	reg, err := ParseRegistry(input)
	if err != nil {
		t.Fatalf("ParseRegistry: %v", err)
	}
	if len(reg.Prefixes) != 0 {
		t.Errorf("Prefixes should be empty for old format, got %v", reg.Prefixes)
	}
	if reg.Projects["exo"] != "/tmp/exo" {
		t.Errorf("Projects[exo] = %q", reg.Projects["exo"])
	}
	if reg.Default != "exo" {
		t.Errorf("Default = %q, want %q", reg.Default, "exo")
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"fn-a001", "fn"},
		{"exo-b002", "exo"},
		{"ko-a001.b002", "ko"},
		{"ko-a001.b002.c003", "ko"},
		{"nohyphen", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractPrefix(tt.id)
		if got != tt.want {
			t.Errorf("extractPrefix(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}


func TestParseRegistryHidden(t *testing.T) {
	input := `projects:
  secret:
    path: /tmp/secret
    hidden: true
  visible:
    path: /tmp/visible
`
	reg, err := ParseRegistry(input)
	if err != nil {
		t.Fatalf("ParseRegistry: %v", err)
	}
	if !reg.Hidden["secret"] {
		t.Error("Hidden[\"secret\"] should be true")
	}
	if reg.Hidden["visible"] {
		t.Error("Hidden[\"visible\"] should be false")
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
