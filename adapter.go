package main

import (
	"os/exec"
	"strings"
)

// AgentAdapter knows how to build an exec.Cmd for a specific agent CLI.
type AgentAdapter interface {
	// BuildCommand returns a ready-to-run Cmd. The caller sets Stdin if needed.
	// prompt is the full prompt text. model may be empty. systemPrompt may be empty.
	BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd
}

// LookupAdapter returns the adapter for a given agent name, or nil if unknown.
func LookupAdapter(name string) AgentAdapter {
	switch name {
	case "claude":
		return &ClaudeAdapter{}
	case "cursor":
		return &CursorAdapter{}
	default:
		return nil
	}
}

// ClaudeAdapter implements AgentAdapter for Claude Code CLI.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd {
	args := []string{"-p", "--output-format", "text"}
	if allowAll {
		args = append(args, "--dangerously-skip-permissions")
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	if systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}

	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(prompt)
	return cmd
}

// CursorAdapter implements AgentAdapter for Cursor's agent CLI.
// Resolves binary as "cursor-agent" or "agent" (whichever is found).
type CursorAdapter struct{}

func (a *CursorAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd {
	// Cursor doesn't support system prompt as a separate flag — inline it.
	fullPrompt := prompt
	if systemPrompt != "" {
		fullPrompt = systemPrompt + "\n\n" + prompt
	}

	args := []string{"-p", fullPrompt, "--output-format", "text"}
	if allowAll {
		args = append(args, "--force")
	}
	// Cursor agent doesn't support --model; skip.

	bin := resolveCursorBin()
	cmd := exec.Command(bin, args...)
	return cmd
}

// resolveCursorBin finds the cursor agent binary.
// Prefers "cursor-agent", falls back to "agent".
func resolveCursorBin() string {
	if path, err := exec.LookPath("cursor-agent"); err == nil {
		return path
	}
	if path, err := exec.LookPath("agent"); err == nil {
		return path
	}
	return "cursor-agent" // will fail at exec time with a clear error
}

// RawCommandAdapter wraps a raw command string (the legacy command: path).
// Behaves like the old hardcoded claude logic for backward compatibility.
type RawCommandAdapter struct {
	Command string
}

func (a *RawCommandAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd {
	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}
	if systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}

	cmd := exec.Command(a.Command, args...)
	cmd.Stdin = strings.NewReader(prompt)
	return cmd
}
