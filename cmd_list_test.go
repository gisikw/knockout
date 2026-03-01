package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectTicketsDir_NoFlag(t *testing.T) {
	// Create a temp dir as the project root
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	// Change to the temp dir so FindTicketsDir works
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	// Call with no --project flag
	resolved, remaining, err := resolveProjectTicketsDir([]string{"arg1", "arg2"})
	if err != nil {
		t.Fatalf("resolveProjectTicketsDir: %v", err)
	}

	if resolved != ticketsDir {
		t.Errorf("resolved = %q, want %q", resolved, ticketsDir)
	}

	if len(remaining) != 2 || remaining[0] != "arg1" || remaining[1] != "arg2" {
		t.Errorf("remaining = %v, want [arg1 arg2]", remaining)
	}
}

func TestResolveProjectTicketsDir_WithFlag(t *testing.T) {
	// Set up a registry with a test project
	regDir := t.TempDir()

	projectDir := t.TempDir()
	projectTicketsDir := filepath.Join(projectDir, ".ko", "tickets")
	os.MkdirAll(projectTicketsDir, 0755)

	// Write a minimal registry
	regContent := `projects:
  testproj: ` + projectDir + `
`

	// Override RegistryPath for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	// Create .config/knockout/projects.yml in the temp home (RegistryPath format)
	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with --project flag
	resolved, remaining, err := resolveProjectTicketsDir([]string{"--project=testproj", "arg1", "arg2"})
	if err != nil {
		t.Fatalf("resolveProjectTicketsDir: %v", err)
	}

	if resolved != projectTicketsDir {
		t.Errorf("resolved = %q, want %q", resolved, projectTicketsDir)
	}

	if len(remaining) != 2 || remaining[0] != "arg1" || remaining[1] != "arg2" {
		t.Errorf("remaining = %v, want [arg1 arg2]", remaining)
	}
}

func TestResolveProjectTicketsDir_UnknownProject(t *testing.T) {
	// Set up a registry with no matching project
	regDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)

	regContent := `projects:
  knownproj: /some/path
`
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with unknown --project
	_, _, err := resolveProjectTicketsDir([]string{"--project=unknown", "arg1"})
	if err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}

	if err.Error() != "unknown project 'unknown'" {
		t.Errorf("error = %q, want 'unknown project 'unknown''", err.Error())
	}
}

func TestResolveProjectTicketsDir_MissingTicketsDir(t *testing.T) {
	// Set up a registry with a project that doesn't have a tickets dir
	regDir := t.TempDir()
	projectDir := t.TempDir()
	// Don't create .ko/tickets

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)

	regContent := `projects:
  testproj: ` + projectDir + `
`
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with --project flag for project without tickets dir
	_, _, err := resolveProjectTicketsDir([]string{"--project=testproj"})
	if err == nil {
		t.Fatal("expected error for missing tickets dir, got nil")
	}

	expectedTicketsDir := filepath.Join(projectDir, ".ko", "tickets")
	expectedErr := "no tickets directory for project 'testproj' (" + expectedTicketsDir + ")"
	if err.Error() != expectedErr {
		t.Errorf("error = %q, want %q", err.Error(), expectedErr)
	}
}

func TestResolveProjectTicketsDir_HashTagShorthand(t *testing.T) {
	// Set up a registry with a test project
	regDir := t.TempDir()

	projectDir := t.TempDir()
	projectTicketsDir := filepath.Join(projectDir, ".ko", "tickets")
	os.MkdirAll(projectTicketsDir, 0755)

	// Write a minimal registry
	regContent := `projects:
  testproj: ` + projectDir + `
`

	// Override RegistryPath for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	// Create .config/knockout/projects.yml in the temp home
	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with #tag shorthand at the beginning
	resolved, remaining, err := resolveProjectTicketsDir([]string{"#testproj", "arg1"})
	if err != nil {
		t.Fatalf("resolveProjectTicketsDir: %v", err)
	}

	if resolved != projectTicketsDir {
		t.Errorf("resolved = %q, want %q", resolved, projectTicketsDir)
	}

	if len(remaining) != 1 || remaining[0] != "arg1" {
		t.Errorf("remaining = %v, want [arg1]", remaining)
	}
}

func TestResolveProjectTicketsDir_HashTagAnyPosition(t *testing.T) {
	// Set up a registry with a test project
	regDir := t.TempDir()

	projectDir := t.TempDir()
	projectTicketsDir := filepath.Join(projectDir, ".ko", "tickets")
	os.MkdirAll(projectTicketsDir, 0755)

	// Write a minimal registry
	regContent := `projects:
  testproj: ` + projectDir + `
`

	// Override RegistryPath for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	// Create .config/knockout/projects.yml in the temp home
	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with #tag shorthand after other args
	resolved, remaining, err := resolveProjectTicketsDir([]string{"arg1", "#testproj"})
	if err != nil {
		t.Fatalf("resolveProjectTicketsDir: %v", err)
	}

	if resolved != projectTicketsDir {
		t.Errorf("resolved = %q, want %q", resolved, projectTicketsDir)
	}

	if len(remaining) != 1 || remaining[0] != "arg1" {
		t.Errorf("remaining = %v, want [arg1]", remaining)
	}
}

func TestResolveProjectTicketsDir_HashTagUnknownProject(t *testing.T) {
	// Set up a registry with no matching project
	regDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)

	regContent := `projects:
  knownproj: /some/path
`
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with unknown #tag
	_, _, err := resolveProjectTicketsDir([]string{"#unknown"})
	if err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}

	if err.Error() != "unknown project 'unknown'" {
		t.Errorf("error = %q, want 'unknown project 'unknown''", err.Error())
	}
}

func TestCmdTriageSet(t *testing.T) {
	ticket := &Ticket{
		ID:       "ko-test",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-01-01T00:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Test Ticket",
	}

	setup := func(t *testing.T) (tmpDir string, ticketsDir string) {
		t.Helper()
		tmpDir = t.TempDir()
		ticketsDir = filepath.Join(tmpDir, ".ko", "tickets")
		if err := os.MkdirAll(ticketsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := SaveTicket(ticketsDir, ticket); err != nil {
			t.Fatal(err)
		}
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir) })
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		return tmpDir, ticketsDir
	}

	t.Run("set triage saves field", func(t *testing.T) {
		_, ticketsDir := setup(t)
		code := cmdTriage([]string{"ko-test", "break this apart"})
		if code != 0 {
			t.Fatalf("cmdTriage returned %d, want 0", code)
		}
		updated, err := LoadTicket(ticketsDir, "ko-test")
		if err != nil {
			t.Fatal(err)
		}
		if updated.Triage != "break this apart" {
			t.Errorf("Triage = %q, want %q", updated.Triage, "break this apart")
		}
	})

	t.Run("multi-word instructions joined", func(t *testing.T) {
		_, ticketsDir := setup(t)
		code := cmdTriage([]string{"ko-test", "unblock", "this", "ticket"})
		if code != 0 {
			t.Fatalf("cmdTriage returned %d, want 0", code)
		}
		updated, err := LoadTicket(ticketsDir, "ko-test")
		if err != nil {
			t.Fatal(err)
		}
		if updated.Triage != "unblock this ticket" {
			t.Errorf("Triage = %q, want %q", updated.Triage, "unblock this ticket")
		}
	})

	t.Run("missing instructions returns non-zero", func(t *testing.T) {
		setup(t)
		code := cmdTriage([]string{"ko-test"})
		if code == 0 {
			t.Error("cmdTriage returned 0, want non-zero")
		}
	})

	t.Run("zero args performs list no error", func(t *testing.T) {
		_, ticketsDir := setup(t)
		// Give the ticket a triage value so it appears in list output
		if err := SaveTicket(ticketsDir, &Ticket{
			ID:       "ko-test",
			Status:   "open",
			Deps:     []string{},
			Created:  "2026-01-01T00:00:00Z",
			Type:     "task",
			Priority: 2,
			Title:    "Test Ticket",
			Triage:   "some triage note",
		}); err != nil {
			t.Fatal(err)
		}
		// Capture stdout to avoid noise
		oldStdout := os.Stdout
		devNull, _ := os.Open(os.DevNull)
		os.Stdout = devNull
		defer func() {
			os.Stdout = oldStdout
			devNull.Close()
		}()
		code := cmdTriage([]string{})
		if code != 0 {
			t.Errorf("cmdTriage(0 args) returned %d, want 0", code)
		}
	})
}

func TestCmdTriageCrossProject(t *testing.T) {
	// Set up a registry with a "fn" prefix project containing a ticket.
	regDir := t.TempDir()
	fnProjectDir := t.TempDir()
	fnTicketsDir := filepath.Join(fnProjectDir, ".ko", "tickets")
	if err := os.MkdirAll(fnTicketsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a ticket in the fn project
	ticket := &Ticket{
		ID:       "fn-test",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-01-01T00:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Cross-project ticket",
	}
	if err := SaveTicket(fnTicketsDir, ticket); err != nil {
		t.Fatal(err)
	}

	// Write registry with fn prefix
	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)
	regContent := "projects:\n  fn: " + fnProjectDir + "\nprefixes:\n  fn: fn\n"
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Override HOME so RegistryPath resolves to our temp registry
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	// Change into a temp dir that is NOT the fn project
	outsideDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
	if err := os.Chdir(outsideDir); err != nil {
		t.Fatal(err)
	}

	code := cmdTriage([]string{"fn-test", "do something"})
	if code != 0 {
		t.Fatalf("cmdTriage cross-project returned %d, want 0", code)
	}

	updated, err := LoadTicket(fnTicketsDir, "fn-test")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Triage != "do something" {
		t.Errorf("Triage = %q, want %q", updated.Triage, "do something")
	}
}

func TestResolveProjectTicketsDir_ProjectFlagOverridesHashTag(t *testing.T) {
	// Set up a registry with two test projects
	regDir := t.TempDir()

	projectDir1 := t.TempDir()
	projectTicketsDir1 := filepath.Join(projectDir1, ".ko", "tickets")
	os.MkdirAll(projectTicketsDir1, 0755)

	projectDir2 := t.TempDir()
	projectTicketsDir2 := filepath.Join(projectDir2, ".ko", "tickets")
	os.MkdirAll(projectTicketsDir2, 0755)

	// Write a minimal registry
	regContent := `projects:
  testproj: ` + projectDir1 + `
  other: ` + projectDir2 + `
`

	// Override RegistryPath for this test
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", regDir)

	// Create .config/knockout/projects.yml in the temp home
	configDir := filepath.Join(regDir, ".config", "knockout")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "projects.yml"), []byte(regContent), 0644)

	// Call with both #tag and --project flag (--project should take precedence)
	resolved, remaining, err := resolveProjectTicketsDir([]string{"#testproj", "--project=other"})
	if err != nil {
		t.Fatalf("resolveProjectTicketsDir: %v", err)
	}

	if resolved != projectTicketsDir2 {
		t.Errorf("resolved = %q, want %q (should use --project not #tag)", resolved, projectTicketsDir2)
	}

	if len(remaining) != 0 {
		t.Errorf("remaining = %v, want []", remaining)
	}
}
