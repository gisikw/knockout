package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangedFilesList(t *testing.T) {
	before := fileSnapshot{
		"main.go":   1000,
		"readme.md": 2000,
		"old.txt":   3000,
	}

	after := fileSnapshot{
		"main.go":    1000, // unchanged
		"readme.md":  2500, // modified
		"newfile.go": 4000, // new
		// old.txt deleted — not in after, so not reported as changed
	}

	changed := changedFilesList(before, after)

	if len(changed) != 2 {
		t.Fatalf("len(changed) = %d, want 2; got %v", len(changed), changed)
	}
	// Sorted alphabetically
	if changed[0] != "newfile.go" {
		t.Errorf("changed[0] = %q, want %q", changed[0], "newfile.go")
	}
	if changed[1] != "readme.md" {
		t.Errorf("changed[1] = %q, want %q", changed[1], "readme.md")
	}
}

func TestChangedFilesListNoChanges(t *testing.T) {
	snap := fileSnapshot{
		"main.go": 1000,
	}
	changed := changedFilesList(snap, snap)
	if len(changed) != 0 {
		t.Errorf("expected no changes, got %v", changed)
	}
}

func TestWriteBackNoteArtifacts(t *testing.T) {
	artifactDir := t.TempDir()
	summaryContent := "## What Changed\n\nAdded a new feature."
	os.WriteFile(filepath.Join(artifactDir, "summary.md"), []byte(summaryContent), 0644)

	ticket := &Ticket{ID: "test-1", Title: "Test ticket"}
	pipeline := &Pipeline{
		Workflows: map[string]*Workflow{
			"task": {
				Name: "task",
				Nodes: []Node{
					{Name: "implement", Type: NodeAction, Prompt: "implement.md"},
					{Name: "review", Type: NodeDecision, Prompt: "review.md", NoteArtifact: "summary.md"},
				},
			},
		},
	}

	writeBackNoteArtifacts(ticket, pipeline, "task", artifactDir)

	if !strings.Contains(ticket.Body, "What Changed") {
		t.Errorf("ticket body should contain artifact content, got: %q", ticket.Body)
	}
	if !strings.Contains(ticket.Body, "Added a new feature.") {
		t.Errorf("ticket body should contain artifact content, got: %q", ticket.Body)
	}
}

func TestWriteBackNoteArtifactsMissing(t *testing.T) {
	artifactDir := t.TempDir()
	// No files created — artifact doesn't exist

	ticket := &Ticket{ID: "test-2", Title: "Test ticket"}
	pipeline := &Pipeline{
		Workflows: map[string]*Workflow{
			"research": {
				Name: "research",
				Nodes: []Node{
					{Name: "investigate", Type: NodeAction, Prompt: "investigate.md", NoteArtifact: "findings.md"},
				},
			},
		},
	}

	writeBackNoteArtifacts(ticket, pipeline, "research", artifactDir)

	if ticket.Body != "" {
		t.Errorf("ticket body should be empty when artifact missing, got: %q", ticket.Body)
	}
}

func TestWriteBackNoteArtifactsRoutedWorkflow(t *testing.T) {
	artifactDir := t.TempDir()
	summaryContent := "## Summary\n\nImplemented the feature."
	os.WriteFile(filepath.Join(artifactDir, "summary.md"), []byte(summaryContent), 0644)

	ticket := &Ticket{ID: "test-routed", Title: "Test routed ticket"}
	// Simulate: main routes to task, task has note_artifact on review.
	// writeBackNoteArtifacts should be called with "task" (the routed workflow),
	// not "main" (which has no note_artifact).
	pipeline := &Pipeline{
		Workflows: map[string]*Workflow{
			"main": {
				Name: "main",
				Nodes: []Node{
					{Name: "classify", Type: NodeDecision, Prompt: "classify.md", Routes: []string{"task"}},
				},
			},
			"task": {
				Name: "task",
				Nodes: []Node{
					{Name: "implement", Type: NodeAction, Prompt: "implement.md"},
					{Name: "review", Type: NodeDecision, Prompt: "review.md", NoteArtifact: "summary.md"},
				},
			},
		},
	}

	// With the bug: finalWorkflow="main", no writeback.
	// With the fix: finalWorkflow="task", writeback happens.
	writeBackNoteArtifacts(ticket, pipeline, "task", artifactDir)

	if !strings.Contains(ticket.Body, "Implemented the feature.") {
		t.Errorf("ticket body should contain routed workflow artifact, got: %q", ticket.Body)
	}

	// Verify it does NOT write back if given "main" (the pre-fix bug)
	ticket2 := &Ticket{ID: "test-routed-2", Title: "Test bug case"}
	writeBackNoteArtifacts(ticket2, pipeline, "main", artifactDir)

	if ticket2.Body != "" {
		t.Errorf("ticket body should be empty when finalWorkflow is main (no note_artifact), got: %q", ticket2.Body)
	}
}

func TestWriteBackNoteArtifactsNoNoteArtifact(t *testing.T) {
	artifactDir := t.TempDir()
	// File exists, but no node declares note_artifact
	os.WriteFile(filepath.Join(artifactDir, "summary.md"), []byte("content"), 0644)

	ticket := &Ticket{ID: "test-3", Title: "Test ticket"}
	pipeline := &Pipeline{
		Workflows: map[string]*Workflow{
			"task": {
				Name: "task",
				Nodes: []Node{
					{Name: "implement", Type: NodeAction, Prompt: "implement.md"},
					{Name: "review", Type: NodeDecision, Prompt: "review.md"},
				},
			},
		},
	}

	writeBackNoteArtifacts(ticket, pipeline, "task", artifactDir)

	if ticket.Body != "" {
		t.Errorf("ticket body should be empty when no note_artifact declared, got: %q", ticket.Body)
	}
}
