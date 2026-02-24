package main

import (
	"os/exec"
	"strings"
)

// AgentAdapter knows how to build an exec.Cmd for a specific agent CLI.
type AgentAdapter interface {
	// BuildCommand returns a ready-to-run Cmd. The caller sets Stdin if needed.
	// prompt is the full prompt text. model may be empty. systemPrompt may be empty.
	// allowedTools may be nil or empty.
	BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd
}

// LookupAdapter returns the adapter for a given agent name, or nil if unknown.
func LookupAdapter(name string) AgentAdapter {
	harness, err := LoadHarness(name)
	if err != nil {
		return nil
	}
	return NewTemplateAdapter(harness)
}

// RawCommandAdapter wraps a raw command string (the legacy command: path).
// Behaves like the old hardcoded claude logic for backward compatibility.
type RawCommandAdapter struct {
	Command string
}

func (a *RawCommandAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd {
	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}
	if systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}
	// allowedTools not implemented for raw command adapter

	cmd := exec.Command(a.Command, args...)
	cmd.Stdin = strings.NewReader(prompt)
	return cmd
}
