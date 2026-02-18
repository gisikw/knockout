package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWritePrefix(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	// No prefix file yet
	if got := ReadPrefix(ticketsDir); got != "" {
		t.Errorf("ReadPrefix on empty dir = %q, want empty", got)
	}

	// Write and read back
	if err := WritePrefix(ticketsDir, "exo"); err != nil {
		t.Fatalf("WritePrefix: %v", err)
	}
	if got := ReadPrefix(ticketsDir); got != "exo" {
		t.Errorf("ReadPrefix after write = %q, want %q", got, "exo")
	}

	// Verify file location
	data, err := os.ReadFile(filepath.Join(dir, ".ko", "prefix"))
	if err != nil {
		t.Fatalf("prefix file not found: %v", err)
	}
	if got := string(data); got != "exo\n" {
		t.Errorf("prefix file content = %q, want %q", got, "exo\n")
	}
}

func TestReadWritePrefixLegacyLayout(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".tickets")
	os.MkdirAll(ticketsDir, 0755)

	// Write and read back with legacy .tickets layout
	if err := WritePrefix(ticketsDir, "leg"); err != nil {
		t.Fatalf("WritePrefix: %v", err)
	}
	if got := ReadPrefix(ticketsDir); got != "leg" {
		t.Errorf("ReadPrefix after write = %q, want %q", got, "leg")
	}
}

func TestDetectPrefixPersists(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	// Create a ticket file so scan finds it
	os.WriteFile(filepath.Join(ticketsDir, "myp-a001.md"), []byte("# Test"), 0644)

	prefix := detectPrefix(ticketsDir)
	if prefix != "myp" {
		t.Fatalf("detectPrefix = %q, want %q", prefix, "myp")
	}

	// Verify it was persisted
	if got := ReadPrefix(ticketsDir); got != "myp" {
		t.Errorf("prefix not persisted: ReadPrefix = %q, want %q", got, "myp")
	}

	// Remove the ticket file — prefix should still come from .ko/prefix
	os.Remove(filepath.Join(ticketsDir, "myp-a001.md"))
	if got := detectPrefix(ticketsDir); got != "myp" {
		t.Errorf("detectPrefix after ticket removal = %q, want %q (should use persisted)", got, "myp")
	}
}

func TestDetectPrefixPersistedWins(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	// Persist one prefix
	WritePrefix(ticketsDir, "abc")

	// Create a ticket with a different prefix
	os.WriteFile(filepath.Join(ticketsDir, "xyz-0001.md"), []byte("# Test"), 0644)

	// Persisted prefix should win
	if got := detectPrefix(ticketsDir); got != "abc" {
		t.Errorf("detectPrefix = %q, want %q (persisted should win over scan)", got, "abc")
	}
}

func TestProjectRoot(t *testing.T) {
	// New layout: .ko/tickets/
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)
	if got := ProjectRoot(ticketsDir); got != dir {
		t.Errorf("ProjectRoot(.ko/tickets) = %q, want %q", got, dir)
	}

	// Legacy layout: .tickets/
	dir2 := t.TempDir()
	ticketsDir2 := filepath.Join(dir2, ".tickets")
	os.MkdirAll(ticketsDir2, 0755)
	if got := ProjectRoot(ticketsDir2); got != dir2 {
		t.Errorf("ProjectRoot(.tickets) = %q, want %q", got, dir2)
	}
}

func TestMigrateTicketsDir(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, ".tickets")
	os.MkdirAll(oldPath, 0755)
	os.WriteFile(filepath.Join(oldPath, "test-0001.md"), []byte("# Test"), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	// FindTicketsDir should migrate .tickets/ to .ko/tickets/
	result, err := FindTicketsDir()
	if err != nil {
		t.Fatalf("FindTicketsDir: %v", err)
	}

	expected := filepath.Join(dir, ".ko", "tickets")
	if result != expected {
		t.Errorf("FindTicketsDir = %q, want %q", result, expected)
	}

	// Old path should be gone
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error(".tickets still exists after migration")
	}

	// Ticket file should be in new location
	if _, err := os.Stat(filepath.Join(expected, "test-0001.md")); err != nil {
		t.Error("ticket file not found in .ko/tickets/ after migration")
	}
}

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
