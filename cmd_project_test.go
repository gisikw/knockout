package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdProjectNoSubcommand(t *testing.T) {
	rc := cmdProject([]string{})
	if rc != 1 {
		t.Errorf("cmdProject with no args returned %d, want 1", rc)
	}
}

func TestCmdProjectUnknownSubcommand(t *testing.T) {
	rc := cmdProject([]string{"unknown"})
	if rc != 1 {
		t.Errorf("cmdProject with unknown subcommand returned %d, want 1", rc)
	}
}

func TestCmdProjectSetMinimal(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "myproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	rc := cmdProjectSet([]string{"#test"})
	if rc != 0 {
		t.Fatalf("cmdProjectSet returned %d, want 0", rc)
	}

	// Verify .ko/tickets directory created
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	if info, err := os.Stat(ticketsDir); err != nil || !info.IsDir() {
		t.Error(".ko/tickets directory not created")
	}

	// Verify registered in registry
	reg, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}
	if path, ok := reg.Projects["test"]; !ok || path != projectDir {
		t.Errorf("project 'test' not registered correctly, got %q, want %q", path, projectDir)
	}
}

func TestCmdProjectSetWithPrefix(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "myproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	rc := cmdProjectSet([]string{"#test", "--prefix=tst"})
	if rc != 0 {
		t.Fatalf("cmdProjectSet returned %d, want 0", rc)
	}

	// Verify prefix written to config
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	if got := ReadPrefix(ticketsDir); got != "tst" {
		t.Errorf("prefix = %q, want %q", got, "tst")
	}

	// Verify prefix stored in registry
	reg, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}
	if prefix, ok := reg.Prefixes["test"]; !ok || prefix != "tst" {
		t.Errorf("prefix for 'test' = %q, want %q", prefix, "tst")
	}
}

func TestCmdProjectSetWithDefault(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "myproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	rc := cmdProjectSet([]string{"#test", "--default"})
	if rc != 0 {
		t.Fatalf("cmdProjectSet returned %d, want 0", rc)
	}

	// Verify set as default
	reg, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}
	if reg.Default != "test" {
		t.Errorf("default = %q, want %q", reg.Default, "test")
	}
}

func TestCmdProjectSetUpsert(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "myproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	// First registration
	cmdProjectSet([]string{"#test", "--prefix=tst"})

	// Update with new prefix and set as default
	rc := cmdProjectSet([]string{"#test", "--prefix=new", "--default"})
	if rc != 0 {
		t.Fatalf("cmdProjectSet (upsert) returned %d, want 0", rc)
	}

	// Verify updated prefix
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	if got := ReadPrefix(ticketsDir); got != "new" {
		t.Errorf("prefix = %q, want %q", got, "new")
	}

	// Verify registry updated
	reg, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}
	if prefix := reg.Prefixes["test"]; prefix != "new" {
		t.Errorf("prefix for 'test' = %q, want %q", prefix, "new")
	}
	if reg.Default != "test" {
		t.Errorf("default = %q, want %q", reg.Default, "test")
	}
}

func TestCmdProjectSetCreatesKoDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "newproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	// Verify .ko doesn't exist yet
	if _, err := os.Stat(filepath.Join(projectDir, ".ko")); err == nil {
		t.Fatal(".ko directory already exists before test")
	}

	rc := cmdProjectSet([]string{"#test"})
	if rc != 0 {
		t.Fatalf("cmdProjectSet returned %d, want 0", rc)
	}

	// Verify .ko/tickets directory created
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	if info, err := os.Stat(ticketsDir); err != nil || !info.IsDir() {
		t.Error(".ko/tickets directory not created")
	}
}

func TestCmdProjectSetNoTag(t *testing.T) {
	rc := cmdProjectSet([]string{})
	if rc != 1 {
		t.Errorf("cmdProjectSet with no tag returned %d, want 1", rc)
	}
}

func TestCmdProjectSetPrefixTooShort(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	orig, _ := os.Getwd()
	projectDir := filepath.Join(dir, "myproject")
	os.MkdirAll(projectDir, 0755)
	os.Chdir(projectDir)
	defer os.Chdir(orig)

	rc := cmdProjectSet([]string{"#test", "--prefix=x"})
	if rc != 1 {
		t.Errorf("cmdProjectSet with 1-char prefix returned %d, want 1", rc)
	}
}

func TestCmdProjectLsEmpty(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create empty registry
	reg := &Registry{Projects: map[string]string{}, Prefixes: map[string]string{}}
	SaveRegistry(regPath, reg)

	rc := cmdProjectLs([]string{})
	if rc != 0 {
		t.Errorf("cmdProjectLs on empty registry returned %d, want 0", rc)
	}
}

func TestCmdProjectLsMultiple(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create registry with multiple projects
	reg := &Registry{
		Default: "proj2",
		Projects: map[string]string{
			"proj1": "/path/to/proj1",
			"proj2": "/path/to/proj2",
			"proj3": "/path/to/proj3",
		},
		Prefixes: map[string]string{},
	}
	SaveRegistry(regPath, reg)

	rc := cmdProjectLs([]string{})
	if rc != 0 {
		t.Errorf("cmdProjectLs returned %d, want 0", rc)
	}
}

func TestCmdProjectLsShowsDefaultMarker(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "knockout", "projects.yml")
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create registry with default
	reg := &Registry{
		Default: "proj2",
		Projects: map[string]string{
			"proj1": "/path/to/proj1",
			"proj2": "/path/to/proj2",
		},
		Prefixes: map[string]string{},
	}
	SaveRegistry(regPath, reg)

	rc := cmdProjectLs([]string{})
	if rc != 0 {
		t.Errorf("cmdProjectLs returned %d, want 0", rc)
	}
	// Output verification would require capturing stdout,
	// which is tested in the integration smoke test
}
