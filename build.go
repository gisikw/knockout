package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Outcome represents the result of a build pipeline.
type Outcome int

const (
	OutcomeSucceed   Outcome = iota
	OutcomeFail
	OutcomeBlocked
	OutcomeDecompose
)

// StageResult holds the output and outcome of a single stage execution.
type StageResult struct {
	Output  string  // stdout from the stage
	Outcome Outcome // parsed outcome (for prompt stages)
	Reason  string  // explanation text after the outcome signal
	BlockOn string  // ticket ID for BLOCKED outcome
	Subtasks []string // subtask titles for DECOMPOSE outcome
}

// ParseStageOutput examines the first line of prompt stage output for
// outcome signals. Pure decision function.
func ParseStageOutput(output string) StageResult {
	lines := strings.SplitN(strings.TrimSpace(output), "\n", 2)
	if len(lines) == 0 {
		return StageResult{Output: output, Outcome: OutcomeSucceed}
	}

	firstLine := strings.TrimSpace(lines[0])
	rest := ""
	if len(lines) > 1 {
		rest = strings.TrimSpace(lines[1])
	}

	switch {
	case firstLine == "FAIL":
		return StageResult{
			Output:  output,
			Outcome: OutcomeFail,
			Reason:  rest,
		}
	case strings.HasPrefix(firstLine, "BLOCKED"):
		parts := strings.Fields(firstLine)
		blockOn := ""
		if len(parts) > 1 {
			blockOn = parts[1]
		}
		return StageResult{
			Output:  output,
			Outcome: OutcomeBlocked,
			BlockOn: blockOn,
			Reason:  rest,
		}
	case firstLine == "DECOMPOSE":
		var subtasks []string
		for _, line := range strings.Split(rest, "\n") {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimSpace(line)
			if line != "" {
				subtasks = append(subtasks, line)
			}
		}
		return StageResult{
			Output:   output,
			Outcome:  OutcomeDecompose,
			Subtasks: subtasks,
		}
	default:
		return StageResult{Output: output, Outcome: OutcomeSucceed}
	}
}

// BuildEligibility checks whether a ticket can be built.
// Pure decision function — returns an error message or "".
func BuildEligibility(t *Ticket, depsResolved bool) string {
	switch t.Status {
	case "in_progress":
		return fmt.Sprintf("ticket '%s' is not eligible for build: already in progress", t.ID)
	case "closed":
		msg := fmt.Sprintf("ticket '%s' is already closed", t.ID)
		// Check for "Closed in:" in notes
		if strings.Contains(t.Body, "Closed in:") {
			// Extract the line
			for _, line := range strings.Split(t.Body, "\n") {
				if strings.Contains(line, "Closed in:") {
					msg += " (" + strings.TrimSpace(line) + ")"
					break
				}
			}
		}
		return msg
	case "blocked":
		return fmt.Sprintf("ticket '%s' is blocked", t.ID)
	case "captured", "routed":
		return fmt.Sprintf("ticket '%s' is not eligible for build: status is '%s'", t.ID, t.Status)
	case "open":
		if !depsResolved {
			return fmt.Sprintf("ticket '%s' has unresolved dependencies", t.ID)
		}
		return ""
	default:
		return fmt.Sprintf("ticket '%s' has unknown status '%s'", t.ID, t.Status)
	}
}

// RunBuild executes the full build pipeline for a ticket.
func RunBuild(ticketsDir string, t *Ticket, p *Pipeline) (Outcome, error) {
	// Create build artifact directory
	buildDir := createBuildDir(ticketsDir, t.ID)

	// Save ticket snapshot
	os.WriteFile(filepath.Join(buildDir, "ticket.md"), []byte(FormatTicket(t)), 0644)

	// Mark ticket as in_progress
	t.Status = "in_progress"
	if err := SaveTicket(ticketsDir, t); err != nil {
		return OutcomeFail, err
	}

	var lastOutput string

	// Execute stages sequentially
	for _, stage := range p.Stages {
		result, err := runStage(ticketsDir, t, p, &stage, lastOutput, buildDir)
		if err != nil {
			// Stage execution error (not an outcome signal)
			applyFailOutcome(ticketsDir, t, stage.Name, err.Error())
			return OutcomeFail, nil
		}

		// Save stage output to build artifacts
		saveStageArtifact(buildDir, stage.Name, result.Output)

		// Check for explicit outcome signals from prompt stages
		if stage.IsPromptStage() && result.Outcome != OutcomeSucceed {
			return applyOutcome(ticketsDir, t, p, &stage, result)
		}

		lastOutput = result.Output
	}

	// All stages passed — run on_succeed hooks
	if err := runHooks(ticketsDir, t, p.OnSucceed, buildDir); err != nil {
		applyFailOutcome(ticketsDir, t, "on_succeed", "on_succeed failed: "+err.Error())
		return OutcomeFail, fmt.Errorf("on_succeed failed")
	}

	// Close ticket
	t.Status = "closed"
	AddNote(t, "ko: SUCCEED")
	if err := SaveTicket(ticketsDir, t); err != nil {
		return OutcomeFail, err
	}

	// Run on_close hooks (ticket already closed — safe even if process dies)
	runHooks(ticketsDir, t, p.OnClose, buildDir)

	return OutcomeSucceed, nil
}

// runStage executes a single stage with retry logic.
func runStage(ticketsDir string, t *Ticket, p *Pipeline, stage *Stage, prevOutput, buildDir string) (StageResult, error) {
	maxAttempts := p.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var result StageResult
		var err error

		if stage.IsPromptStage() {
			result, err = runPromptStage(ticketsDir, t, p, stage, prevOutput)
		} else if stage.IsRunStage() {
			result, err = runRunStage(stage)
		} else {
			return StageResult{}, fmt.Errorf("stage '%s' has neither prompt nor run", stage.Name)
		}

		if err != nil {
			// Execution error — retry if attempts remain
			if attempt+1 < maxAttempts {
				continue
			}
			return StageResult{}, fmt.Errorf("stage '%s' failed after %d attempts: %v", stage.Name, maxAttempts, err)
		}

		// Check for explicit signals (never retry these)
		if stage.IsPromptStage() && result.Outcome != OutcomeSucceed {
			return result, nil
		}

		return result, nil
	}

	return StageResult{}, fmt.Errorf("stage '%s' failed after %d attempts", stage.Name, maxAttempts)
}

// runPromptStage invokes the configured command with ticket context.
func runPromptStage(ticketsDir string, t *Ticket, p *Pipeline, stage *Stage, prevOutput string) (StageResult, error) {
	promptContent, err := LoadPromptFile(ticketsDir, stage.Prompt)
	if err != nil {
		return StageResult{}, err
	}

	// Build the full prompt
	var prompt strings.Builder
	prompt.WriteString("## Ticket\n\n")
	prompt.WriteString(fmt.Sprintf("# %s\n", t.Title))
	if t.Body != "" {
		prompt.WriteString(t.Body)
	}
	prompt.WriteString("\n\n")

	if prevOutput != "" {
		prompt.WriteString("## Previous Stage Output\n\n")
		prompt.WriteString(prevOutput)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString(fmt.Sprintf("## Discretion Level: %s\n\n", p.Discretion))
	prompt.WriteString(DiscretionGuidance(p.Discretion))
	prompt.WriteString("\n\n")

	prompt.WriteString("## Instructions\n\n")
	prompt.WriteString(promptContent)

	// Determine model
	model := p.Model
	if stage.Model != "" {
		model = stage.Model
	}

	// Invoke the command
	output, err := invokeCommand(p.Command, model, prompt.String())
	if err != nil {
		return StageResult{}, err
	}

	return ParseStageOutput(output), nil
}

// invokeCommand runs the configured LLM command with the given prompt.
func invokeCommand(command, model, prompt string) (string, error) {
	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command '%s' failed: %v", command, err)
	}
	return string(out), nil
}

// runRunStage executes a shell command.
func runRunStage(stage *Stage) (StageResult, error) {
	cmd := exec.Command("sh", "-c", stage.Run)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return StageResult{}, fmt.Errorf("command failed: %s\n%s", err, string(out))
	}
	return StageResult{
		Output:  string(out),
		Outcome: OutcomeSucceed,
	}, nil
}

// applyOutcome handles non-SUCCEED outcomes from a stage.
func applyOutcome(ticketsDir string, t *Ticket, p *Pipeline, stage *Stage, result StageResult) (Outcome, error) {
	switch result.Outcome {
	case OutcomeFail:
		applyFailOutcome(ticketsDir, t, stage.Name, result.Reason)
		return OutcomeFail, nil

	case OutcomeBlocked:
		if result.BlockOn != "" {
			// Resolve the blocking ticket ID
			blockID, err := ResolveID(ticketsDir, result.BlockOn)
			if err != nil {
				applyFailOutcome(ticketsDir, t, stage.Name, fmt.Sprintf("BLOCKED on '%s' but ticket not found", result.BlockOn))
				return OutcomeFail, nil
			}
			t.Deps = append(t.Deps, blockID)
		}
		t.Status = "open" // back to open, but now has deps so won't be in ready
		note := fmt.Sprintf("ko: BLOCKED at stage '%s'", stage.Name)
		if result.Reason != "" {
			note += " — " + result.Reason
		}
		AddNote(t, note)
		SaveTicket(ticketsDir, t)
		return OutcomeBlocked, nil

	case OutcomeDecompose:
		return applyDecompose(ticketsDir, t, p, stage, result)

	default:
		return OutcomeSucceed, nil
	}
}

// applyDecompose creates child tickets and blocks the parent on them.
func applyDecompose(ticketsDir string, t *Ticket, p *Pipeline, stage *Stage, result StageResult) (Outcome, error) {
	// Check depth guard
	depth := Depth(t.ID)
	if depth >= p.MaxDepth {
		t.Status = "blocked"
		AddNote(t, fmt.Sprintf("ko: DECOMPOSE denied — max decomposition depth (%d) reached", p.MaxDepth))
		SaveTicket(ticketsDir, t)
		return OutcomeFail, nil
	}

	// Create child tickets
	var childIDs []string
	for _, subtask := range result.Subtasks {
		child := NewChildTicket(t.ID, subtask)
		if err := SaveTicket(ticketsDir, child); err != nil {
			applyFailOutcome(ticketsDir, t, stage.Name, "failed to create child ticket: "+err.Error())
			return OutcomeFail, nil
		}
		childIDs = append(childIDs, child.ID)
	}

	// Block parent on children
	t.Deps = append(t.Deps, childIDs...)
	t.Status = "open" // open but blocked by deps
	AddNote(t, fmt.Sprintf("ko: DECOMPOSE — created %d children: %s", len(childIDs), strings.Join(childIDs, ", ")))
	SaveTicket(ticketsDir, t)

	return OutcomeDecompose, nil
}

// applyFailOutcome marks a ticket as blocked with a failure note.
func applyFailOutcome(ticketsDir string, t *Ticket, stageName, reason string) {
	t.Status = "blocked"
	note := fmt.Sprintf("ko: FAIL at stage '%s'", stageName)
	if reason != "" {
		note += " — " + reason
	}
	AddNote(t, note)
	SaveTicket(ticketsDir, t)
}

// runHooks executes a list of shell commands with env vars set.
func runHooks(ticketsDir string, t *Ticket, hooks []string, buildDir string) error {
	if len(hooks) == 0 {
		return nil
	}

	projectRoot := filepath.Dir(ticketsDir)

	// Read CHANGED_FILES if it exists
	changedFiles := ""
	cfPath := filepath.Join(buildDir, "changed_files.txt")
	if data, err := os.ReadFile(cfPath); err == nil {
		changedFiles = strings.TrimSpace(string(data))
	}

	for _, hook := range hooks {
		// Expand variables
		expanded := os.Expand(hook, func(key string) string {
			switch key {
			case "TICKET_ID":
				return t.ID
			case "CHANGED_FILES":
				return changedFiles
			default:
				return os.Getenv(key)
			}
		})

		cmd := exec.Command("sh", "-c", expanded)
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(),
			"TICKET_ID="+t.ID,
			"CHANGED_FILES="+changedFiles,
		)

		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("hook '%s' failed: %v\n%s", hook, err, string(out))
		}
	}
	return nil
}

// createBuildDir creates a timestamped build artifact directory.
func createBuildDir(ticketsDir, ticketID string) string {
	projectRoot := filepath.Dir(ticketsDir)
	ts := time.Now().UTC().Format("20060102-150405")
	dir := filepath.Join(projectRoot, ".ko", "builds", ts+"-"+ticketID)
	os.MkdirAll(dir, 0755)
	return dir
}

// saveStageArtifact writes stage output to the build directory.
func saveStageArtifact(buildDir, stageName, output string) {
	os.WriteFile(filepath.Join(buildDir, stageName+".md"), []byte(output), 0644)
}
