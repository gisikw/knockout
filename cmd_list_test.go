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
