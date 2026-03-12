package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdShowQuestionsSection(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	ticket := `---
id: test-0001
status: blocked
deps: []
created: 2026-03-11T10:00:00Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: Tabs or spaces?
    options:
      - label: Spaces
        value: spaces
      - label: Tabs
        value: tabs
  - id: q2
    question: Which library?
    context: INVARIANTS.md says no external deps
    options:
      - label: Standard library
        value: stdlib
        description: Matches invariant
      - label: External library
        value: external
        description: Violates invariant
---
# Blocked ticket
`
	os.WriteFile(filepath.Join(ticketsDir, "test-0001.md"), []byte(ticket), 0644)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := cmdShow([]string{"test-0001"})

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Fatalf("cmdShow() = %d, want 0; output: %s", exitCode, output)
	}

	// Verify questions section is present
	if !strings.Contains(output, "## Questions") {
		t.Errorf("output missing ## Questions section:\n%s", output)
	}
	if !strings.Contains(output, "? Tabs or spaces? [q1]") {
		t.Errorf("output missing first question:\n%s", output)
	}
	if !strings.Contains(output, "? Which library? [q2]") {
		t.Errorf("output missing second question:\n%s", output)
	}
	if !strings.Contains(output, "Context: INVARIANTS.md says no external deps") {
		t.Errorf("output missing question context:\n%s", output)
	}
	if !strings.Contains(output, "  - Standard library — Matches invariant") {
		t.Errorf("output missing option with description:\n%s", output)
	}
	if !strings.Contains(output, "  - Spaces") {
		t.Errorf("output missing option without description:\n%s", output)
	}
}

func TestCmdShowNoQuestionsWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	ticket := `---
id: test-0002
status: open
deps: []
created: 2026-03-11T10:00:00Z
type: task
priority: 2
---
# Normal ticket
`
	os.WriteFile(filepath.Join(ticketsDir, "test-0002.md"), []byte(ticket), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	exitCode := cmdShow([]string{"test-0002"})

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Fatalf("cmdShow() = %d, want 0", exitCode)
	}

	if strings.Contains(output, "## Questions") {
		t.Errorf("output should not have Questions section for ticket without questions:\n%s", output)
	}
}
