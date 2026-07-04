package main

import "testing"

func TestParseQQLMapping(t *testing.T) {
	content := `
# comment line
default_realm: knockout-legacy
projects:
  fort-nix:
    realm: fort-nix
    campaign: fort-nix-maintenance
  questbook:
    realm: questbook
    campaign: questbook-buildout
  gee:
    realm: gee
`
	m := ParseQQLMapping(content)

	if m.DefaultRealm != "knockout-legacy" {
		t.Errorf("default_realm = %q, want knockout-legacy", m.DefaultRealm)
	}
	if len(m.Projects) != 3 {
		t.Fatalf("got %d projects, want 3: %+v", len(m.Projects), m.Projects)
	}
	if pm := m.Projects["fort-nix"]; pm.Realm != "fort-nix" || pm.Campaign != "fort-nix-maintenance" {
		t.Errorf("fort-nix = %+v", pm)
	}
	if pm := m.Projects["gee"]; pm.Realm != "gee" || pm.Campaign != "" {
		t.Errorf("gee = %+v, want realm=gee campaign empty", pm)
	}
}

func TestQQLMappingResolve(t *testing.T) {
	m := &QQLMapping{
		DefaultRealm: "fallback-realm",
		Projects: map[string]QQLProjectMap{
			"fort-nix": {Realm: "fort-nix", Campaign: "fn-maint"},
			"noreal":   {Campaign: "only-campaign"},
		},
	}

	tests := []struct {
		tag          string
		wantRealm    string
		wantCampaign string
	}{
		{"fort-nix", "fort-nix", "fn-maint"},
		{"noreal", "fallback-realm", "only-campaign"}, // realm empty → default
		{"unmapped", "fallback-realm", ""},            // unknown → default realm
	}
	for _, tc := range tests {
		r, c := m.Resolve(tc.tag)
		if r != tc.wantRealm || c != tc.wantCampaign {
			t.Errorf("Resolve(%q) = (%q,%q), want (%q,%q)", tc.tag, r, c, tc.wantRealm, tc.wantCampaign)
		}
	}
}

func TestQQLMappingResolveNoDefaultFallsBackToTag(t *testing.T) {
	m := &QQLMapping{Projects: map[string]QQLProjectMap{}}
	r, c := m.Resolve("some-project")
	if r != "some-project" || c != "" {
		t.Errorf("Resolve = (%q,%q), want (some-project,'')", r, c)
	}
}
