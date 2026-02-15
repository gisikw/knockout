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
		Prefixes: map[string]string{
			"exo":      "exo",
			"fort-nix": "fn",
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
	if len(parsed.Prefixes) != len(reg.Prefixes) {
		t.Fatalf("len(Prefixes) = %d, want %d", len(parsed.Prefixes), len(reg.Prefixes))
	}
	for k, v := range reg.Prefixes {
		if parsed.Prefixes[k] != v {
			t.Errorf("Prefixes[%s] = %q, want %q", k, parsed.Prefixes[k], v)
		}
	}
}

func TestParseRegistryBackwardCompatible(t *testing.T) {
	// Old format without prefixes section should still parse
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

func TestParseTags(t *testing.T) {
	tests := []struct {
		input    string
		wantTitle string
		wantTags  []string
	}{
		{"Fix the bug", "Fix the bug", nil},
		{"Do thing #fort-nix", "Do thing", []string{"fort-nix"}},
		{"Refactor #fort-nix #security #urgent", "Refactor", []string{"fort-nix", "security", "urgent"}},
		{"No tags here", "No tags here", nil},
		{"#solo", "", []string{"solo"}},
	}
	for _, tt := range tests {
		title, tags := ParseTags(tt.input)
		if title != tt.wantTitle {
			t.Errorf("ParseTags(%q) title = %q, want %q", tt.input, title, tt.wantTitle)
		}
		if len(tags) != len(tt.wantTags) {
			t.Errorf("ParseTags(%q) tags = %v, want %v", tt.input, tags, tt.wantTags)
			continue
		}
		for i := range tags {
			if tags[i] != tt.wantTags[i] {
				t.Errorf("ParseTags(%q) tags[%d] = %q, want %q", tt.input, i, tags[i], tt.wantTags[i])
			}
		}
	}
}

func TestRouteTicket(t *testing.T) {
	reg := &Registry{
		Default: "exo",
		Projects: map[string]string{
			"fort-nix": "/projects/fort-nix",
			"exo":      "/projects/exo",
		},
	}

	// No tag — local
	d := RouteTicket("Fix bug", reg, "/projects/local")
	if d.TargetPath != "/projects/local" || d.Status != "open" || d.IsRouted {
		t.Errorf("no tag: got target=%q status=%q routed=%v", d.TargetPath, d.Status, d.IsRouted)
	}

	// Recognized tag — routed
	d = RouteTicket("Add thing #fort-nix", reg, "/projects/local")
	if d.TargetPath != "/projects/fort-nix" || d.Status != "routed" || !d.IsRouted {
		t.Errorf("recognized tag: got target=%q status=%q routed=%v", d.TargetPath, d.Status, d.IsRouted)
	}
	if d.Title != "Add thing" {
		t.Errorf("recognized tag: title = %q, want %q", d.Title, "Add thing")
	}

	// Unrecognized tag — captured to default
	d = RouteTicket("Build page #marketing", reg, "/projects/local")
	if d.TargetPath != "/projects/exo" || d.Status != "captured" || !d.IsCaptured {
		t.Errorf("unrecognized tag: got target=%q status=%q captured=%v", d.TargetPath, d.Status, d.IsCaptured)
	}
	if d.RoutingTag != "marketing" {
		t.Errorf("unrecognized tag: routingTag = %q, want %q", d.RoutingTag, "marketing")
	}

	// Multiple tags — first routes, rest are labels
	d = RouteTicket("Refactor auth #fort-nix #security #urgent", reg, "/projects/local")
	if d.RoutingTag != "fort-nix" {
		t.Errorf("multi tag: routingTag = %q", d.RoutingTag)
	}
	if len(d.ExtraTags) != 2 || d.ExtraTags[0] != "security" || d.ExtraTags[1] != "urgent" {
		t.Errorf("multi tag: extraTags = %v", d.ExtraTags)
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
