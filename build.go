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

// BuildEligibility checks whether a ticket can be built.
// Pure decision function — returns an error message or "".
func BuildEligibility(t *Ticket, depsResolved bool) string {
	switch t.Status {
	case "in_progress":
		return fmt.Sprintf("ticket '%s' is not eligible for build: already in progress", t.ID)
	case "closed":
		msg := fmt.Sprintf("ticket '%s' is already closed", t.ID)
		if strings.Contains(t.Body, "Closed in:") {
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
	buildDir := createBuildDir(ticketsDir, t.ID)

	// Save ticket snapshot
	os.WriteFile(filepath.Join(buildDir, "ticket.md"), []byte(FormatTicket(t)), 0644)

	// Create workspace
	wsDir, err := CreateWorkspace(buildDir)
	if err != nil {
		return OutcomeFail, fmt.Errorf("failed to create workspace: %v", err)
	}

	// Mark ticket as in_progress
	t.Status = "in_progress"
	if err := SaveTicket(ticketsDir, t); err != nil {
		return OutcomeFail, err
	}

	// Snapshot project files before stages run
	projectRoot := filepath.Dir(ticketsDir)
	beforeSnapshot := snapshotFiles(projectRoot)

	// Visit counters: node name -> visit count
	visits := make(map[string]int)

	// Execute starting from "main" workflow
	outcome, err := runWorkflow(ticketsDir, t, p, "main", visits, wsDir, buildDir)
	if err != nil {
		applyFailOutcome(ticketsDir, t, "build", err.Error())
		return OutcomeFail, nil
	}

	if outcome != OutcomeSucceed {
		return outcome, nil
	}

	// Compute changed files
	writeChangedFiles(buildDir, projectRoot, beforeSnapshot)

	// All workflows passed — run on_succeed hooks
	if err := runHooks(ticketsDir, t, p.OnSucceed, buildDir, wsDir); err != nil {
		applyFailOutcome(ticketsDir, t, "on_succeed", "on_succeed failed: "+err.Error())
		return OutcomeFail, fmt.Errorf("on_succeed failed")
	}

	// Close ticket
	t.Status = "closed"
	AddNote(t, "ko: SUCCEED")
	if err := SaveTicket(ticketsDir, t); err != nil {
		return OutcomeFail, err
	}

	// Run on_close hooks (ticket already closed)
	runHooks(ticketsDir, t, p.OnClose, buildDir, wsDir)

	return OutcomeSucceed, nil
}

// runWorkflow executes a single workflow, following route dispositions to other
// workflows. Returns the terminal outcome.
func runWorkflow(ticketsDir string, t *Ticket, p *Pipeline, wfName string, visits map[string]int, wsDir, buildDir string) (Outcome, error) {
	wf, ok := p.Workflows[wfName]
	if !ok {
		return OutcomeFail, fmt.Errorf("unknown workflow '%s'", wfName)
	}

	for i := 0; i < len(wf.Nodes); i++ {
		node := &wf.Nodes[i]

		// Check visit limit
		visits[node.Name]++
		if visits[node.Name] > node.MaxVisits {
			applyFailOutcome(ticketsDir, t, node.Name,
				fmt.Sprintf("node '%s' exceeded max_visits (%d)", node.Name, node.MaxVisits))
			return OutcomeFail, nil
		}

		// Resolve model: node > workflow > pipeline
		model := resolveModel(p, wf, node)

		// Execute the node
		output, err := runNode(ticketsDir, t, p, node, model, wsDir)
		if err != nil {
			applyFailOutcome(ticketsDir, t, node.Name, err.Error())
			return OutcomeFail, nil
		}

		// Tee output to workspace
		TeeOutput(wsDir, wfName, node.Name, output)

		// Save to build artifacts
		saveStageArtifact(buildDir, wfName+"."+node.Name, output)

		// Action nodes: output isn't parsed, just continue
		if node.Type == NodeAction {
			continue
		}

		// Decision nodes: parse disposition
		disp, err := extractDisposition(output)
		if err != nil {
			// Should not happen — retries already exhausted in runNode
			applyFailOutcome(ticketsDir, t, node.Name, err.Error())
			return OutcomeFail, nil
		}

		outcome, err := applyDisposition(ticketsDir, t, p, node, wfName, disp, visits, wsDir, buildDir)
		if err != nil {
			return OutcomeFail, err
		}
		if outcome != OutcomeSucceed {
			return outcome, nil
		}
		// OutcomeSucceed from applyDisposition means "continue" — next node
	}

	// Reached end of workflow = succeed
	return OutcomeSucceed, nil
}

// runNode executes a single node with retry logic.
func runNode(ticketsDir string, t *Ticket, p *Pipeline, node *Node, model, wsDir string) (string, error) {
	maxAttempts := p.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var output string
		var err error

		if node.IsPromptNode() {
			output, err = runPromptNode(ticketsDir, t, p, node, model, wsDir)
		} else if node.IsRunNode() {
			output, err = runRunNode(node, wsDir)
		} else {
			return "", fmt.Errorf("node '%s' has neither prompt nor run", node.Name)
		}

		if err != nil {
			if attempt+1 < maxAttempts {
				continue
			}
			return "", fmt.Errorf("node '%s' failed after %d attempts: %v", node.Name, maxAttempts, err)
		}

		// For decision nodes, validate disposition extraction
		if node.Type == NodeDecision {
			if _, extractErr := extractDisposition(output); extractErr != nil {
				if attempt+1 < maxAttempts {
					continue // retry on invalid disposition
				}
				return "", fmt.Errorf("node '%s' failed after %d attempts: %v", node.Name, maxAttempts, extractErr)
			}
		}

		return output, nil
	}

	return "", fmt.Errorf("node '%s' failed after %d attempts", node.Name, maxAttempts)
}

// extractDisposition parses the disposition from decision node output.
func extractDisposition(output string) (Disposition, error) {
	jsonStr, ok := ExtractLastFencedJSON(output)
	if !ok {
		return Disposition{}, fmt.Errorf("no fenced JSON block found in output")
	}
	return ParseDisposition(jsonStr)
}

// applyDisposition handles a parsed disposition from a decision node.
// Returns OutcomeSucceed for "continue" (advance to next node).
func applyDisposition(ticketsDir string, t *Ticket, p *Pipeline, node *Node, currentWF string, disp Disposition, visits map[string]int, wsDir, buildDir string) (Outcome, error) {
	switch disp.Type {
	case "continue":
		return OutcomeSucceed, nil

	case "fail":
		applyFailOutcome(ticketsDir, t, node.Name, disp.Reason)
		return OutcomeFail, nil

	case "blocked":
		return applyBlockedDisposition(ticketsDir, t, node, disp)

	case "decompose":
		return applyDecomposeDisposition(ticketsDir, t, p, node, disp)

	case "route":
		if !contains(node.Routes, disp.Workflow) {
			applyFailOutcome(ticketsDir, t, node.Name,
				fmt.Sprintf("node '%s' tried to route to '%s' but only declares routes: %v",
					node.Name, disp.Workflow, node.Routes))
			return OutcomeFail, nil
		}
		return runWorkflow(ticketsDir, t, p, disp.Workflow, visits, wsDir, buildDir)

	default:
		applyFailOutcome(ticketsDir, t, node.Name, fmt.Sprintf("unknown disposition '%s'", disp.Type))
		return OutcomeFail, nil
	}
}

// applyBlockedDisposition wires a dependency and returns the ticket to open.
func applyBlockedDisposition(ticketsDir string, t *Ticket, node *Node, disp Disposition) (Outcome, error) {
	if disp.BlockOn != "" {
		blockID, err := ResolveID(ticketsDir, disp.BlockOn)
		if err != nil {
			applyFailOutcome(ticketsDir, t, node.Name,
				fmt.Sprintf("BLOCKED on '%s' but ticket not found", disp.BlockOn))
			return OutcomeFail, nil
		}
		t.Deps = append(t.Deps, blockID)
	}
	t.Status = "open"
	note := fmt.Sprintf("ko: BLOCKED at node '%s'", node.Name)
	if disp.Reason != "" {
		note += " — " + disp.Reason
	}
	AddNote(t, note)
	SaveTicket(ticketsDir, t)
	return OutcomeBlocked, nil
}

// applyDecomposeDisposition creates child tickets and blocks the parent.
func applyDecomposeDisposition(ticketsDir string, t *Ticket, p *Pipeline, node *Node, disp Disposition) (Outcome, error) {
	depth := Depth(t.ID)
	if depth >= p.MaxDepth {
		t.Status = "blocked"
		AddNote(t, fmt.Sprintf("ko: DECOMPOSE denied — max decomposition depth (%d) reached", p.MaxDepth))
		SaveTicket(ticketsDir, t)
		return OutcomeFail, nil
	}

	var childIDs []string
	for _, subtask := range disp.Subtasks {
		child := NewChildTicket(t.ID, subtask)
		if err := SaveTicket(ticketsDir, child); err != nil {
			applyFailOutcome(ticketsDir, t, node.Name, "failed to create child ticket: "+err.Error())
			return OutcomeFail, nil
		}
		childIDs = append(childIDs, child.ID)
	}

	t.Deps = append(t.Deps, childIDs...)
	t.Status = "open"
	AddNote(t, fmt.Sprintf("ko: DECOMPOSE — created %d children: %s", len(childIDs), strings.Join(childIDs, ", ")))
	SaveTicket(ticketsDir, t)

	return OutcomeDecompose, nil
}

// runPromptNode invokes the configured command with ticket context.
func runPromptNode(ticketsDir string, t *Ticket, p *Pipeline, node *Node, model, wsDir string) (string, error) {
	promptContent, err := LoadPromptFile(ticketsDir, node.Prompt)
	if err != nil {
		return "", err
	}

	var prompt strings.Builder
	prompt.WriteString("## Ticket\n\n")
	prompt.WriteString(fmt.Sprintf("# %s\n", t.Title))
	if t.Body != "" {
		prompt.WriteString(t.Body)
	}
	prompt.WriteString("\n\n")

	prompt.WriteString(fmt.Sprintf("## Discretion Level: %s\n\n", p.Discretion))
	prompt.WriteString(DiscretionGuidance(p.Discretion))
	prompt.WriteString("\n\n")

	prompt.WriteString("## Instructions\n\n")
	prompt.WriteString(promptContent)

	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}

	// Decision nodes get the disposition schema injected
	if node.Type == NodeDecision {
		args = append(args, "--append-system-prompt", DispositionSchema)
	}

	cmd := exec.Command(p.Command, args...)
	cmd.Stdin = strings.NewReader(prompt.String())
	cmd.Env = append(os.Environ(), "KO_TICKET_WORKSPACE="+wsDir)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command '%s' failed: %v", p.Command, err)
	}
	return string(out), nil
}

// runRunNode executes a shell command.
func runRunNode(node *Node, wsDir string) (string, error) {
	cmd := exec.Command("sh", "-c", node.Run)
	cmd.Env = append(os.Environ(), "KO_TICKET_WORKSPACE="+wsDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s\n%s", err, string(out))
	}
	return string(out), nil
}

// resolveModel returns the most specific model override for a node.
func resolveModel(p *Pipeline, wf *Workflow, node *Node) string {
	if node.Model != "" {
		return node.Model
	}
	if wf.Model != "" {
		return wf.Model
	}
	return p.Model
}

// applyFailOutcome marks a ticket as blocked with a failure note.
func applyFailOutcome(ticketsDir string, t *Ticket, nodeName, reason string) {
	t.Status = "blocked"
	note := fmt.Sprintf("ko: FAIL at node '%s'", nodeName)
	if reason != "" {
		note += " — " + reason
	}
	AddNote(t, note)
	SaveTicket(ticketsDir, t)
}

// runHooks executes a list of shell commands with env vars set.
func runHooks(ticketsDir string, t *Ticket, hooks []string, buildDir, wsDir string) error {
	if len(hooks) == 0 {
		return nil
	}

	projectRoot := filepath.Dir(ticketsDir)

	changedFiles := ""
	cfPath := filepath.Join(buildDir, "changed_files.txt")
	if data, err := os.ReadFile(cfPath); err == nil {
		changedFiles = strings.TrimSpace(string(data))
	}

	for _, hook := range hooks {
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
			"KO_TICKET_WORKSPACE="+wsDir,
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
func saveStageArtifact(buildDir, name, output string) {
	os.WriteFile(filepath.Join(buildDir, name+".md"), []byte(output), 0644)
}

// fileSnapshot records mod times for files in a project directory.
type fileSnapshot map[string]int64

func snapshotFiles(projectRoot string) fileSnapshot {
	snap := make(fileSnapshot)
	filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".ko" || base == ".tickets" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return nil
		}
		snap[rel] = info.ModTime().UnixNano()
		return nil
	})
	return snap
}

func changedFilesList(before, after fileSnapshot) []string {
	var changed []string
	for path, modTime := range after {
		prevMod, existed := before[path]
		if !existed || modTime != prevMod {
			changed = append(changed, path)
		}
	}
	sortStrings(changed)
	return changed
}

func writeChangedFiles(buildDir, projectRoot string, before fileSnapshot) {
	after := snapshotFiles(projectRoot)
	changed := changedFilesList(before, after)
	if len(changed) > 0 {
		content := strings.Join(changed, "\n")
		os.WriteFile(filepath.Join(buildDir, "changed_files.txt"), []byte(content), 0644)
	}
}

// contains checks if a slice contains a string.
func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
