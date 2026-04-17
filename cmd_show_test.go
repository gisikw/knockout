package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdShowQuestionsSection(t *testing.T) {
	defer setupTestDB(t)()
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	ticket := &Ticket{
		ID:       "test-0001",
		Status:   "blocked",
		Deps:     []string{},
		Created:  "2026-03-11T10:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Blocked ticket",
		PlanQuestions: []PlanQuestion{
			{
				ID:       "q1",
				Question: "Tabs or spaces?",
				Options: []QuestionOption{
					{Label: "Spaces", Value: "spaces"},
					{Label: "Tabs", Value: "tabs"},
				},
			},
			{
				ID:       "q2",
				Question: "Which library?",
				Context:  "INVARIANTS.md says no external deps",
				Options: []QuestionOption{
					{Label: "Standard library", Value: "stdlib", Description: "Matches invariant"},
					{Label: "External library", Value: "external", Description: "Violates invariant"},
				},
			},
		},
	}
	if err := SaveTicket(ticketsDir, ticket); err != nil {
		t.Fatal(err)
	}

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
	defer setupTestDB(t)()
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, ".ko", "tickets")
	os.MkdirAll(ticketsDir, 0755)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	ticket := &Ticket{
		ID:       "test-0002",
		Status:   "open",
		Deps:     []string{},
		Created:  "2026-03-11T10:00:00Z",
		Type:     "task",
		Priority: 2,
		Title:    "Normal ticket",
	}
	if err := SaveTicket(ticketsDir, ticket); err != nil {
		t.Fatal(err)
	}

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
