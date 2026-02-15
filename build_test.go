package main

import (
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
