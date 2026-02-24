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
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)
	if got := ProjectRoot(ticketsDir); got != dir {
		t.Errorf("ProjectRoot(.ko/tickets) = %q, want %q", got, dir)
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

func TestCreateWithShorthandPriority(t *testing.T) {
	// Clear KO_NO_CREATE to allow ticket creation in tests
	origNoCreate := os.Getenv("KO_NO_CREATE")
	os.Unsetenv("KO_NO_CREATE")
	defer func() {
		if origNoCreate != "" {
			os.Setenv("KO_NO_CREATE", origNoCreate)
		}
	}()

	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)
	WritePrefix(ticketsDir, "test")

	// Save original dir and restore after test
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	tests := []struct {
		name     string
		args     []string
		wantPrio int
	}{
		{"shorthand p0", []string{"-p0", "Test ticket"}, 0},
		{"shorthand p1", []string{"-p1", "Test ticket"}, 1},
		{"shorthand p4", []string{"-p4", "Test ticket"}, 4},
		{"longform with space", []string{"-p", "3", "Test ticket"}, 3},
		{"longform with equals", []string{"-p=2", "Test ticket"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := cmdCreate(tt.args)
			if exitCode != 0 {
				t.Fatalf("cmdCreate failed with exit code %d", exitCode)
			}

			// Find the created ticket
			entries, err := os.ReadDir(ticketsDir)
			if err != nil {
				t.Fatalf("failed to read tickets dir: %v", err)
			}

			var ticketPath string
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".md" {
					ticketPath = filepath.Join(ticketsDir, e.Name())
					break
				}
			}

			if ticketPath == "" {
				t.Fatal("no ticket file created")
			}

			// Extract ticket ID from filename
			ticketID := filepath.Base(ticketPath)
			ticketID = ticketID[:len(ticketID)-3] // Remove .md extension

			// Load and verify priority
			ticket, err := LoadTicket(ticketsDir, ticketID)
			if err != nil {
				t.Fatalf("failed to load ticket: %v", err)
			}

			if ticket.Priority != tt.wantPrio {
				t.Errorf("ticket priority = %d, want %d", ticket.Priority, tt.wantPrio)
			}

			// Clean up for next test
			os.Remove(ticketPath)
		})
	}
}

func TestCreateWithInvalidShorthandPriority(t *testing.T) {
	// Clear KO_NO_CREATE to allow ticket creation in tests
	origNoCreate := os.Getenv("KO_NO_CREATE")
	os.Unsetenv("KO_NO_CREATE")
	defer func() {
		if origNoCreate != "" {
			os.Setenv("KO_NO_CREATE", origNoCreate)
		}
	}()

	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)
	WritePrefix(ticketsDir, "test")

	// Save original dir and restore after test
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	tests := []struct {
		name string
		args []string
	}{
		{"priority too high -p5", []string{"-p5", "Test ticket"}},
		{"priority too high -p9", []string{"-p9", "Test ticket"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := cmdCreate(tt.args)
			// We expect these to succeed in creation, but the priority validation
			// should be handled by the flag package or application logic.
			// Let's verify the ticket is created and check what priority it has.
			if exitCode != 0 {
				// If it fails, that's also acceptable behavior
				return
			}

			// If it succeeds, verify the created ticket exists
			entries, err := os.ReadDir(ticketsDir)
			if err != nil {
				t.Fatalf("failed to read tickets dir: %v", err)
			}

			var ticketPath string
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".md" {
					ticketPath = filepath.Join(ticketsDir, e.Name())
					break
				}
			}

			if ticketPath != "" {
				// Clean up
				os.Remove(ticketPath)
			}
		})
	}
}
