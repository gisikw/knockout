package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadHarness_BuiltInClaude(t *testing.T) {
	config, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}
	if config.ScriptPath == "" {
		t.Error("expected script path to be non-empty")
	}
}

func TestLoadHarness_BuiltInCursor(t *testing.T) {
	config, err := LoadHarness("cursor")
	if err != nil {
		t.Fatalf("LoadHarness(cursor) failed: %v", err)
	}
	if config.ScriptPath == "" {
		t.Error("expected script path to be non-empty")
	}
}

func TestLoadHarness_UnknownName(t *testing.T) {
	_, err := LoadHarness("nonexistent")
	if err == nil {
		t.Error("expected error for unknown harness, got nil")
	}
}

func TestLoadHarness_UserOverride(t *testing.T) {
	// Create a temp directory for user config
	tmpDir := t.TempDir()
	harnessDir := filepath.Join(tmpDir, ".config", "knockout", "agent-harnesses")
	if err := os.MkdirAll(harnessDir, 0755); err != nil {
		t.Fatalf("failed to create harness dir: %v", err)
	}

	// Write a custom shell harness
	customHarness := `#!/bin/sh
echo "custom agent"
`
	customPath := filepath.Join(harnessDir, "custom")
	if err := os.WriteFile(customPath, []byte(customHarness), 0755); err != nil {
		t.Fatalf("failed to write custom harness: %v", err)
	}

	// Set HOME to temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load the custom harness
	config, err := LoadHarness("custom")
	if err != nil {
		t.Fatalf("LoadHarness(custom) failed: %v", err)
	}
	if config.ScriptPath != customPath {
		t.Errorf("expected script path=%q, got %q", customPath, config.ScriptPath)
	}
}

func TestLoadHarness_ProjectOverride(t *testing.T) {
	// Save current dir and restore after test
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temp directory for project
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	harnessDir := filepath.Join(".ko", "agent-harnesses")
	if err := os.MkdirAll(harnessDir, 0755); err != nil {
		t.Fatalf("failed to create harness dir: %v", err)
	}

	// Write a project-local shell harness that overrides claude
	projectHarness := `#!/bin/sh
echo "project claude"
`
	projectPath := filepath.Join(harnessDir, "claude")
	if err := os.WriteFile(projectPath, []byte(projectHarness), 0755); err != nil {
		t.Fatalf("failed to write project harness: %v", err)
	}

	// Load should get the project override
	config, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}
	if !strings.HasSuffix(config.ScriptPath, "claude") {
		t.Errorf("expected script path to end with 'claude', got %q", config.ScriptPath)
	}
}


func TestShellAdapter_BuildCommand(t *testing.T) {
	// Create a test shell script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-harness.sh")
	script := `#!/bin/sh
echo "test"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	adapter := NewShellAdapter(scriptPath)
	cmd := adapter.BuildCommand("test prompt", "sonnet", "test system", true, []string{"read", "write"})

	// Check that command points to the script
	if cmd.Path != scriptPath {
		t.Errorf("expected cmd.Path=%q, got %q", scriptPath, cmd.Path)
	}

	// Check that KO_* environment variables are set
	envMap := make(map[string]string)
	for _, env := range cmd.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["KO_PROMPT"] != "test prompt" {
		t.Errorf("expected KO_PROMPT='test prompt', got %q", envMap["KO_PROMPT"])
	}
	if envMap["KO_MODEL"] != "sonnet" {
		t.Errorf("expected KO_MODEL='sonnet', got %q", envMap["KO_MODEL"])
	}
	if envMap["KO_SYSTEM_PROMPT"] != "test system" {
		t.Errorf("expected KO_SYSTEM_PROMPT='test system', got %q", envMap["KO_SYSTEM_PROMPT"])
	}
	if envMap["KO_ALLOW_ALL"] != "true" {
		t.Errorf("expected KO_ALLOW_ALL='true', got %q", envMap["KO_ALLOW_ALL"])
	}
	if envMap["KO_ALLOWED_TOOLS"] != "read,write" {
		t.Errorf("expected KO_ALLOWED_TOOLS='read,write', got %q", envMap["KO_ALLOWED_TOOLS"])
	}
}

func TestShellAdapter_BuildCommandWithoutOptionalParams(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-harness.sh")
	script := `#!/bin/sh
echo "test"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	adapter := NewShellAdapter(scriptPath)
	cmd := adapter.BuildCommand("test prompt", "", "", false, nil)

	envMap := make(map[string]string)
	for _, env := range cmd.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["KO_MODEL"] != "" {
		t.Errorf("expected KO_MODEL to be empty string, got %q", envMap["KO_MODEL"])
	}
	if envMap["KO_SYSTEM_PROMPT"] != "" {
		t.Errorf("expected KO_SYSTEM_PROMPT to be empty string, got %q", envMap["KO_SYSTEM_PROMPT"])
	}
	if envMap["KO_ALLOW_ALL"] != "false" {
		t.Errorf("expected KO_ALLOW_ALL='false', got %q", envMap["KO_ALLOW_ALL"])
	}
	if envMap["KO_ALLOWED_TOOLS"] != "" {
		t.Errorf("expected KO_ALLOWED_TOOLS to be empty string, got %q", envMap["KO_ALLOWED_TOOLS"])
	}
}

