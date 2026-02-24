package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadHarness_BuiltInClaude(t *testing.T) {
	h, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}
	if h.Binary != "claude" {
		t.Errorf("expected binary=claude, got %q", h.Binary)
	}
	if len(h.Args) == 0 {
		t.Error("expected args to be non-empty")
	}
}

func TestLoadHarness_BuiltInCursor(t *testing.T) {
	h, err := LoadHarness("cursor")
	if err != nil {
		t.Fatalf("LoadHarness(cursor) failed: %v", err)
	}
	if len(h.BinaryFallbacks) == 0 {
		t.Error("expected binary_fallbacks to be non-empty")
	}
	if h.BinaryFallbacks[0] != "cursor-agent" {
		t.Errorf("expected first fallback=cursor-agent, got %q", h.BinaryFallbacks[0])
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

	// Write a custom harness
	customHarness := `binary: custom-agent
args:
  - "--custom-flag"
`
	customPath := filepath.Join(harnessDir, "custom.yaml")
	if err := os.WriteFile(customPath, []byte(customHarness), 0644); err != nil {
		t.Fatalf("failed to write custom harness: %v", err)
	}

	// Set HOME to temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load the custom harness
	h, err := LoadHarness("custom")
	if err != nil {
		t.Fatalf("LoadHarness(custom) failed: %v", err)
	}
	if h.Binary != "custom-agent" {
		t.Errorf("expected binary=custom-agent, got %q", h.Binary)
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

	// Write a project-local harness that overrides claude
	projectHarness := `binary: project-claude
args:
  - "--project-flag"
`
	projectPath := filepath.Join(harnessDir, "claude.yaml")
	if err := os.WriteFile(projectPath, []byte(projectHarness), 0644); err != nil {
		t.Fatalf("failed to write project harness: %v", err)
	}

	// Load should get the project override
	h, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}
	if h.Binary != "project-claude" {
		t.Errorf("expected binary=project-claude (project override), got %q", h.Binary)
	}
}

func TestTemplateAdapter_ClaudeCommand(t *testing.T) {
	h, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}

	adapter := NewTemplateAdapter(h)
	cmd := adapter.BuildCommand("test prompt", "sonnet", "test system", true, nil)

	if cmd.Path != "claude" && !strings.HasSuffix(cmd.Path, "/claude") {
		t.Errorf("expected claude binary, got %q", cmd.Path)
	}

	// Check that args contain expected flags
	args := cmd.Args
	hasModel := false
	hasSystemPrompt := false
	hasDangerouslySkip := false
	hasOutputFormat := false

	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) && args[i+1] == "sonnet" {
			hasModel = true
		}
		if arg == "--append-system-prompt" && i+1 < len(args) && args[i+1] == "test system" {
			hasSystemPrompt = true
		}
		if arg == "--dangerously-skip-permissions" {
			hasDangerouslySkip = true
		}
		if arg == "--output-format" {
			hasOutputFormat = true
		}
	}

	if !hasModel {
		t.Error("expected --model sonnet in args")
	}
	if !hasSystemPrompt {
		t.Error("expected --append-system-prompt in args")
	}
	if !hasDangerouslySkip {
		t.Error("expected --dangerously-skip-permissions in args")
	}
	if !hasOutputFormat {
		t.Error("expected --output-format in args")
	}

	// Check stdin is set
	if cmd.Stdin == nil {
		t.Error("expected stdin to be set for claude")
	}
}

func TestTemplateAdapter_ClaudeCommandWithoutOptionalFlags(t *testing.T) {
	h, err := LoadHarness("claude")
	if err != nil {
		t.Fatalf("LoadHarness(claude) failed: %v", err)
	}

	adapter := NewTemplateAdapter(h)
	cmd := adapter.BuildCommand("test prompt", "", "", false, nil)

	// Check that optional flags are NOT present
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "--model") {
		t.Error("expected --model to be omitted when empty")
	}
	if strings.Contains(args, "--append-system-prompt") {
		t.Error("expected --append-system-prompt to be omitted when empty")
	}
	if strings.Contains(args, "--dangerously-skip-permissions") {
		t.Error("expected --dangerously-skip-permissions to be omitted when allowAll=false")
	}
}

func TestTemplateAdapter_CursorCommand(t *testing.T) {
	h, err := LoadHarness("cursor")
	if err != nil {
		t.Fatalf("LoadHarness(cursor) failed: %v", err)
	}

	adapter := NewTemplateAdapter(h)
	cmd := adapter.BuildCommand("test prompt", "sonnet", "test system", true, nil)

	// Check binary resolution (should be cursor-agent or agent)
	bin := filepath.Base(cmd.Path)
	if bin != "cursor-agent" && bin != "agent" {
		t.Errorf("expected cursor-agent or agent binary, got %q", bin)
	}

	// Check that args contain expected flags
	args := cmd.Args
	hasModel := false
	hasForce := false
	hasOutputFormat := false
	hasPromptWithSystem := false

	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) && args[i+1] == "sonnet" {
			hasModel = true
		}
		if arg == "--force" {
			hasForce = true
		}
		if arg == "--output-format" {
			hasOutputFormat = true
		}
		// Check that prompt includes system prompt (inlined)
		if arg == "-p" && i+1 < len(args) {
			nextArg := args[i+1]
			if strings.Contains(nextArg, "test system") && strings.Contains(nextArg, "test prompt") {
				hasPromptWithSystem = true
			}
		}
	}

	if !hasModel {
		t.Error("expected --model sonnet in args")
	}
	if !hasForce {
		t.Error("expected --force in args")
	}
	if !hasOutputFormat {
		t.Error("expected --output-format in args")
	}
	if !hasPromptWithSystem {
		t.Error("expected inlined system prompt in cursor prompt arg")
	}

	// Cursor should NOT use stdin
	if cmd.Stdin != nil {
		t.Error("expected stdin to be nil for cursor (prompt passed as arg)")
	}
}

func TestTemplateAdapter_BinaryFallbackResolution(t *testing.T) {
	// Create a harness with fallbacks
	h := &Harness{
		BinaryFallbacks: []string{"nonexistent-binary-1", "nonexistent-binary-2"},
		Args:            []string{"-p"},
	}

	adapter := NewTemplateAdapter(h)
	cmd := adapter.BuildCommand("test", "", "", false, nil)

	// Should fall back to first option
	if cmd.Path != "nonexistent-binary-1" {
		t.Errorf("expected first fallback binary, got %q", cmd.Path)
	}
}
