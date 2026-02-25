package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdBuildInit_createsFiles(t *testing.T) {
	dir := t.TempDir()

	// Create .tickets so findProjectRoot works
	os.MkdirAll(filepath.Join(dir, ".tickets"), 0755)

	// Run from the temp dir
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	code := cmdAgentInit(nil)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	// Verify files exist
	for _, path := range []string{
		".ko/config.yaml",
		".ko/prompts/triage.md",
		".ko/prompts/implement.md",
		".ko/prompts/review.md",
		".ko/.gitignore",
	} {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected %s to exist: %v", path, err)
		}
	}

	// Verify .ko/.gitignore contains agent runtime files
	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".ko", ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .ko/.gitignore: %v", err)
	}
	gitignoreContent := string(gitignoreData)
	for _, entry := range []string{"agent.lock", "agent.pid", "agent.log", "agent.heartbeat"} {
		if !strings.Contains(gitignoreContent, entry) {
			t.Errorf("expected .ko/.gitignore to contain %q", entry)
		}
	}

	// Verify config.yaml is parseable
	config, err := LoadConfig(filepath.Join(dir, ".ko", "config.yaml"))
	if err != nil {
		t.Fatalf("generated config.yaml failed to parse: %v", err)
	}
	if _, ok := config.Pipeline.Workflows["main"]; !ok {
		t.Error("expected 'main' workflow in generated config")
	}
	if config.Project.Prefix != "ko" {
		t.Errorf("expected default prefix 'ko', got %q", config.Project.Prefix)
	}
}

func TestCmdBuildInit_refusesOverwrite(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".tickets"), 0755)
	os.MkdirAll(filepath.Join(dir, ".ko"), 0755)
	os.WriteFile(filepath.Join(dir, ".ko", "config.yaml"), []byte("existing"), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	code := cmdAgentInit(nil)
	if code != 1 {
		t.Fatalf("expected exit 1 (already exists), got %d", code)
	}

	// Verify existing file wasn't overwritten
	data, _ := os.ReadFile(filepath.Join(dir, ".ko", "config.yaml"))
	if string(data) != "existing" {
		t.Error("existing config.yaml was overwritten")
	}
}

func TestCmdBuildInit_worksWithoutTicketsDir(t *testing.T) {
	dir := t.TempDir()

	// No .tickets dir — should use cwd
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	code := cmdAgentInit(nil)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if _, err := os.Stat(filepath.Join(dir, ".ko", "config.yaml")); err != nil {
		t.Error("expected config.yaml to be created in cwd")
	}
}
