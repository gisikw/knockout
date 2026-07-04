package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCmdExport(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Isolated registry via XDG_CONFIG_HOME.
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	// A project with a couple of tickets in its DB-backed store.
	projectDir := t.TempDir()
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatal(err)
	}

	parent := &Ticket{
		ID: "tp-0001", Title: "Parent ticket", Status: "open", Type: "task",
		Priority: 1, Deps: []string{}, Created: "2026-01-01T00:00:00Z",
		Body: "parent body", Tags: []string{"alpha", "beta"},
	}
	child := &Ticket{
		ID: "tp-0002", Title: "Child ticket", Status: "closed", Type: "bug",
		Priority: 2, Deps: []string{"tp-0001"}, Created: "2026-01-02T00:00:00Z",
		Body: "child body",
	}
	if err := SaveTicket(ticketsDir, parent); err != nil {
		t.Fatal(err)
	}
	if err := SaveTicket(ticketsDir, child); err != nil {
		t.Fatal(err)
	}
	// One history event on the parent.
	EmitMutationEvent(ticketsDir, "tp-0001", "create", map[string]interface{}{"title": "Parent ticket"})

	// Registry pointing at the project.
	regDir := filepath.Join(configHome, "knockout")
	if err := os.MkdirAll(regDir, 0755); err != nil {
		t.Fatal(err)
	}
	reg := "projects:\n  testproj:\n    path: " + projectDir + "\n    prefix: tp\n    default: true\n"
	if err := os.WriteFile(filepath.Join(regDir, "projects.yml"), []byte(reg), 0644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "export.json")
	if code := cmdExport([]string{"--out", outPath}); code != 0 {
		t.Fatalf("cmdExport returned %d", code)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var exp KnockoutExport
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatalf("export not valid JSON: %v", err)
	}

	if exp.SchemaVersion != ExportSchemaVersion {
		t.Errorf("schema_version = %q, want %q", exp.SchemaVersion, ExportSchemaVersion)
	}
	if exp.ProjectCount != 1 || exp.TicketCount != 2 {
		t.Fatalf("counts: projects=%d tickets=%d, want 1/2", exp.ProjectCount, exp.TicketCount)
	}
	pj := exp.Projects[0]
	if pj.Tag != "testproj" || pj.Prefix != "tp" || !pj.IsDefault {
		t.Errorf("project meta = %+v", pj)
	}
	if pj.Path != projectDir {
		t.Errorf("project path = %q, want %q", pj.Path, projectDir)
	}

	byID := map[string]ExportTicket{}
	for _, tk := range pj.Tickets {
		byID[tk.ID] = tk
	}
	p := byID["tp-0001"]
	if p.Title != "Parent ticket" || p.Status != "open" || p.Priority != 1 {
		t.Errorf("parent fields = %+v", p)
	}
	if len(p.Tags) != 2 {
		t.Errorf("parent tags = %v, want 2", p.Tags)
	}
	if len(p.History) != 1 || p.History[0].EventType != "create" {
		t.Errorf("parent history = %+v, want one create event", p.History)
	}
	c := byID["tp-0002"]
	if len(c.Deps) != 1 || c.Deps[0] != "tp-0001" {
		t.Errorf("child deps = %v, want [tp-0001]", c.Deps)
	}
	// deps/tags must always be present arrays (never null) for the import contract.
	if p.Deps == nil || c.Tags == nil {
		t.Errorf("deps/tags must be non-nil arrays: parent.Deps=%v child.Tags=%v", p.Deps, c.Tags)
	}
}

func TestCmdExportUnknownProject(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	regDir := filepath.Join(configHome, "knockout")
	os.MkdirAll(regDir, 0755)
	os.WriteFile(filepath.Join(regDir, "projects.yml"), []byte("projects:\n"), 0644)

	if code := cmdExport([]string{"--project", "does-not-exist", "--out", filepath.Join(t.TempDir(), "x.json")}); code == 0 {
		t.Error("cmdExport with unknown project returned 0, want nonzero")
	}
}
