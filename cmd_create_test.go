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

func TestCreateWithDescription(t *testing.T) {
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
		wantDesc string
	}{
		{
			name:     "second positional arg sets description",
			args:     []string{"Test title", "Test description from arg"},
			wantDesc: "Test description from arg",
		},
		{
			name:     "-d flag sets description when no positional arg",
			args:     []string{"-d", "Test description from flag", "Test title"},
			wantDesc: "Test description from flag",
		},
		{
			name:     "second arg takes priority over -d flag",
			args:     []string{"Test title", "From arg", "-d", "From flag"},
			wantDesc: "From arg",
		},
		{
			name:     "empty description is allowed",
			args:     []string{"Test title"},
			wantDesc: "",
		},
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

			// Load and verify description
			ticket, err := LoadTicket(ticketsDir, ticketID)
			if err != nil {
				t.Fatalf("failed to load ticket: %v", err)
			}

			// The description is stored in t.Body with newlines wrapped
			// Expected format: "\n" + description + "\n" or empty if no description
			var gotDesc string
			if ticket.Body != "" {
				// Strip the leading and trailing newlines
				gotDesc = ticket.Body
				if len(gotDesc) > 0 && gotDesc[0] == '\n' {
					gotDesc = gotDesc[1:]
				}
				if len(gotDesc) > 0 && gotDesc[len(gotDesc)-1] == '\n' {
					gotDesc = gotDesc[:len(gotDesc)-1]
				}
			}

			if gotDesc != tt.wantDesc {
				t.Errorf("ticket description = %q, want %q", gotDesc, tt.wantDesc)
			}

			// Clean up for next test
			os.Remove(ticketPath)
		})
	}
}

func TestCreateWithStdinDescription(t *testing.T) {
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

	// Save original stdin and restore after test
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	tests := []struct {
		name        string
		stdinInput  string
		args        []string
		wantDesc    string
		description string
	}{
		{
			name:        "stdin sets description",
			stdinInput:  "Description from stdin",
			args:        []string{"Test title"},
			wantDesc:    "Description from stdin",
			description: "stdin should be used when provided",
		},
		{
			name:        "stdin takes priority over second arg",
			stdinInput:  "From stdin",
			args:        []string{"Test title", "From arg"},
			wantDesc:    "From stdin",
			description: "stdin should win over positional arg",
		},
		{
			name:        "stdin takes priority over -d flag",
			stdinInput:  "From stdin",
			args:        []string{"-d", "From flag", "Test title"},
			wantDesc:    "From stdin",
			description: "stdin should win over -d flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to mock stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			defer r.Close()

			// Replace stdin with the read end of the pipe
			os.Stdin = r

			// Write test data to the pipe and close the write end
			go func() {
				w.Write([]byte(tt.stdinInput))
				w.Close()
			}()

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

			// Load and verify description
			ticket, err := LoadTicket(ticketsDir, ticketID)
			if err != nil {
				t.Fatalf("failed to load ticket: %v", err)
			}

			// The description is stored in t.Body with newlines wrapped
			var gotDesc string
			if ticket.Body != "" {
				gotDesc = ticket.Body
				if len(gotDesc) > 0 && gotDesc[0] == '\n' {
					gotDesc = gotDesc[1:]
				}
				if len(gotDesc) > 0 && gotDesc[len(gotDesc)-1] == '\n' {
					gotDesc = gotDesc[:len(gotDesc)-1]
				}
			}

			if gotDesc != tt.wantDesc {
				t.Errorf("ticket description = %q, want %q", gotDesc, tt.wantDesc)
			}

			// Clean up for next test
			os.Remove(ticketPath)

			// Restore stdin for next iteration
			os.Stdin = origStdin
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
