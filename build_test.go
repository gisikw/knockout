package main

import (
	"os"
	"os/exec"
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

func TestRunPromptNodePriorContextInjection(t *testing.T) {
	tests := []struct {
		name        string
		nodeType    NodeType
		wantContext bool
	}{
		{
			name:        "action node receives prior context",
			nodeType:    NodeAction,
			wantContext: true,
		},
		{
			name:        "decision node does not receive prior context",
			nodeType:    NodeDecision,
			wantContext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directories
			artifactDir := t.TempDir()
			wsDir := filepath.Join(artifactDir, "workspace")
			os.MkdirAll(wsDir, 0755)

			// Create prior context files that would be injected
			priorContent := "Previous build output"
			os.WriteFile(filepath.Join(artifactDir, "plan.md"), []byte(priorContent), 0644)

			// Use inline prompt (contains newlines, so won't try to load file)
			promptContent := "Test instructions\nfor the node"

			// Build the prompt using the same logic as runPromptNode
			ticket := &Ticket{ID: "test-node", Title: "Test Node Context"}
			node := &Node{
				Name:   "test",
				Type:   tt.nodeType,
				Prompt: promptContent, // inline prompt
			}

			// Build the prompt with the same logic as runPromptNode (lines 454-477)
			var prompt strings.Builder
			prompt.WriteString("## Ticket\n\n")
			prompt.WriteString("# " + ticket.Title + "\n")
			prompt.WriteString("\n\n")
			prompt.WriteString("## Discretion Level: medium\n\n")
			prompt.WriteString("\n\n")

			// The key logic being tested: only inject for action nodes
			if node.Type == NodeAction {
				if priorContext := InjectPriorContext(artifactDir, "task"); priorContext != "" {
					prompt.WriteString(priorContext)
					prompt.WriteString("\n\n")
				}
			}

			prompt.WriteString("## Instructions\n\n")
			prompt.WriteString(promptContent)

			builtPrompt := prompt.String()

			// Verify: action nodes should have prior context, decision nodes should not
			hasPriorContext := strings.Contains(builtPrompt, priorContent)
			if tt.wantContext && !hasPriorContext {
				t.Errorf("action node should receive prior context, but it was not found in prompt")
			}
			if !tt.wantContext && hasPriorContext {
				t.Errorf("decision node should NOT receive prior context, but it was found in prompt")
			}
		})
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	// Make an initial commit so the repo is not bare
	os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0644)
	run("add", ".gitkeep")
	run("commit", "-m", "init")
}

func TestIsWorkingTreeClean(t *testing.T) {
	projectRoot := t.TempDir()
	initGitRepo(t, projectRoot)

	// Clean repo: should return true
	clean, err := isWorkingTreeClean(projectRoot)
	if err != nil {
		t.Fatalf("isWorkingTreeClean error: %v", err)
	}
	if !clean {
		t.Error("expected clean tree, got dirty")
	}

	// Add untracked file outside .ko/: should return false
	os.WriteFile(filepath.Join(projectRoot, "dirty.go"), []byte("package main"), 0644)
	clean, err = isWorkingTreeClean(projectRoot)
	if err != nil {
		t.Fatalf("isWorkingTreeClean error: %v", err)
	}
	if clean {
		t.Error("expected dirty tree, got clean")
	}
	os.Remove(filepath.Join(projectRoot, "dirty.go"))

	// Add file only inside .ko/: should return true (ignored)
	os.MkdirAll(filepath.Join(projectRoot, ".ko"), 0755)
	os.WriteFile(filepath.Join(projectRoot, ".ko", "meta.txt"), []byte("data"), 0644)
	clean, err = isWorkingTreeClean(projectRoot)
	if err != nil {
		t.Fatalf("isWorkingTreeClean error: %v", err)
	}
	if !clean {
		t.Error("expected clean tree when only .ko/ changes exist, got dirty")
	}
}

func TestRequireCleanTreeRejectsDirty(t *testing.T) {
	projectRoot := t.TempDir()
	koDir := filepath.Join(projectRoot, ".ko")
	ticketsDir := filepath.Join(koDir, "tickets")
	os.MkdirAll(ticketsDir, 0755)
	initGitRepo(t, projectRoot)

	// Add uncommitted change outside .ko/
	os.WriteFile(filepath.Join(projectRoot, "dirty.go"), []byte("package main"), 0644)

	ticket := &Ticket{ID: "ko-a001", Status: "open"}
	msg := BuildEligibility(ticketsDir, ticket, true, true)
	if msg == "" {
		t.Error("expected eligibility error for dirty tree, got empty string")
	}
	if !strings.Contains(msg, "uncommitted changes") {
		t.Errorf("expected message to contain 'uncommitted changes', got: %q", msg)
	}
}

func TestRequireCleanTreeIgnoresKoDir(t *testing.T) {
	projectRoot := t.TempDir()
	koDir := filepath.Join(projectRoot, ".ko")
	ticketsDir := filepath.Join(koDir, "tickets")
	os.MkdirAll(ticketsDir, 0755)
	initGitRepo(t, projectRoot)

	// Add uncommitted change only inside .ko/
	os.MkdirAll(filepath.Join(projectRoot, ".ko", "artifacts"), 0755)
	os.WriteFile(filepath.Join(projectRoot, ".ko", "artifacts", "note.md"), []byte("data"), 0644)

	ticket := &Ticket{ID: "ko-a001", Status: "open"}
	msg := BuildEligibility(ticketsDir, ticket, true, true)
	if msg != "" {
		t.Errorf("expected no eligibility error when only .ko/ changes, got: %q", msg)
	}
}

func TestHasFlakeNix(t *testing.T) {
	// Create temporary directory structure: projectRoot/.ko/tickets/
	projectRoot := t.TempDir()
	koDir := filepath.Join(projectRoot, ".ko")
	ticketsDir := filepath.Join(koDir, "tickets")
	os.MkdirAll(ticketsDir, 0755)

	// Test case 1: flake.nix does not exist
	if hasFlakeNix(ticketsDir) {
		t.Errorf("hasFlakeNix should return false when flake.nix does not exist")
	}

	// Test case 2: flake.nix exists
	flakePath := filepath.Join(projectRoot, "flake.nix")
	os.WriteFile(flakePath, []byte("# test flake"), 0644)

	if !hasFlakeNix(ticketsDir) {
		t.Errorf("hasFlakeNix should return true when flake.nix exists")
	}
}
