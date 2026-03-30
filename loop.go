package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoopConfig holds the parameters for a loop run.
type LoopConfig struct {
	MaxTickets  int           // max tickets to process (0 = unlimited)
	MaxDuration time.Duration // max wall-clock duration (0 = unlimited)
	Quiet       bool          // suppress per-ticket stdout output
	Verbose     bool          // stream full agent output to stdout
}

// LoopResult summarizes the outcome of a loop run.
type LoopResult struct {
	Processed int // total tickets attempted
	Succeeded int // tickets that reached SUCCEED
	Failed    int // tickets that reached FAIL
	Blocked   int // tickets that reached BLOCKED
	Decomposed int // tickets that reached DECOMPOSE
	Stopped   string // why the loop stopped: "empty", "max_tickets", "max_duration", "build_error"
}

// ShouldContinue decides whether the loop should process another ticket.
// Pure decision function.
func (c *LoopConfig) ShouldContinue(processed int, elapsed time.Duration) (bool, string) {
	if c.MaxTickets > 0 && processed >= c.MaxTickets {
		return false, "max_tickets"
	}
	if c.MaxDuration > 0 && elapsed >= c.MaxDuration {
		return false, "max_duration"
	}
	return true, ""
}

// TriageQueue returns all tickets with a non-empty triage field,
// sorted by priority then modified. Pure query.
func TriageQueue(ticketsDir string) ([]*Ticket, error) {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return nil, err
	}

	var triageable []*Ticket
	for _, t := range tickets {
		if t.Triage != "" {
			triageable = append(triageable, t)
		}
	}
	SortByPriorityThenModified(triageable)
	return triageable, nil
}

// runTriagePass runs triage on all tickets in the triage queue.
// Logs failures but continues processing. Respects the stop channel between
// runs. Returns count of tickets successfully triaged.
func runTriagePass(ticketsDir string, p *Pipeline, verbose bool, quiet bool, stop <-chan struct{}) int {
	queue, err := TriageQueue(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop: triage queue error: %v\n", err)
		return 0
	}

	count := 0
	for _, t := range queue {
		// Check for stop signal between triage runs
		if stop != nil {
			select {
			case <-stop:
				return count
			default:
			}
		}

		if !quiet {
			fmt.Printf("loop: triaging %s — %s\n", t.ID, t.Title)
		}
		if err := runAgentTriage(ticketsDir, t, p, verbose); err != nil {
			fmt.Fprintf(os.Stderr, "loop: triage failed for %s: %v\n", t.ID, err)
			continue
		}
		count++
	}
	return count
}

// ReadyQueue returns the IDs of tickets ready to build, sorted by priority.
func ReadyQueue(ticketsDir string) ([]string, error) {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return nil, err
	}

	var ready []*Ticket
	for _, t := range tickets {
		if IsReady(t.Status, AllDepsResolved(ticketsDir, t.Deps)) && !IsSnoozed(t.Snooze, time.Now()) && t.Triage == "" {
			ready = append(ready, t)
		}
	}
	SortByPriorityThenModified(ready)

	ids := make([]string, len(ready))
	for i, t := range ready {
		ids[i] = t.ID
	}
	return ids, nil
}

// RunLoop burns down the ready queue, building one ticket at a time (or
// multiple in parallel when Pipeline.Workers > 1).
// Sets KO_NO_CREATE to prevent spawned agents from creating tickets.
// If stop is non-nil, the loop checks it between builds and exits with
// "signal" when closed.
func RunLoop(ticketsDir string, p *Pipeline, config LoopConfig, log *EventLogger, stop <-chan struct{}) LoopResult {
	os.Setenv("KO_NO_CREATE", "1")
	defer os.Unsetenv("KO_NO_CREATE")

	start := time.Now()
	result := LoopResult{}

	if p.Workers > 1 {
		pruneWorktrees(ticketsDir)
	}

	for {
		// Check limits
		if ok, reason := config.ShouldContinue(result.Processed, time.Since(start)); !ok {
			result.Stopped = reason
			return result
		}

		// Run triage pre-pass before picking up ready tickets
		runTriagePass(ticketsDir, p, config.Verbose, config.Quiet, stop)

		// Get next ready ticket — empty queue is definitive, check before signal
		queue, err := ReadyQueue(ticketsDir)
		if err != nil {
			result.Stopped = "build_error"
			return result
		}
		if len(queue) == 0 {
			result.Stopped = "empty"
			return result
		}

		// Check for stop signal (only meaningful if there's work remaining)
		if stop != nil {
			select {
			case <-stop:
				result.Stopped = "signal"
				return result
			default:
			}
		}

		if p.Workers <= 1 {
			// === Sequential path (original behavior) ===
			id := queue[0]
			t, err := LoadTicket(ticketsDir, id)
			if err != nil {
				result.Stopped = "build_error"
				return result
			}

			if !config.Quiet {
				fmt.Printf("loop: building %s — %s\n", id, t.Title)
			}
			log.LoopTicketStart(id, t.Title)

			outcome, err := RunBuild(ticketsDir, t, p, log, config.Verbose)
			result.Processed++

			if err != nil {
				fmt.Fprintf(os.Stderr, "loop: build error for %s: %v\n", id, err)
				result.Stopped = "build_error"
				return result
			}

			switch outcome {
			case OutcomeSucceed:
				result.Succeeded++
			case OutcomeFail:
				result.Failed++
			case OutcomeBlocked:
				result.Blocked++
			case OutcomeDecompose:
				result.Decomposed++
			}

			outcomeStr := outcomeString(outcome)
			if !config.Quiet {
				fmt.Printf("loop: %s %s\n", id, strings.ToUpper(outcomeStr))
			}
			log.LoopTicketComplete(id, outcomeStr)
		} else {
			// === Parallel path (worktree-isolated) ===
			batchSize := p.Workers
			if len(queue) < batchSize {
				batchSize = len(queue)
			}
			if config.MaxTickets > 0 {
				remaining := config.MaxTickets - result.Processed
				if remaining <= 0 {
					result.Stopped = "max_tickets"
					return result
				}
				if batchSize > remaining {
					batchSize = remaining
				}
			}

			batch := queue[:batchSize]

			type buildResult struct {
				id         string
				title      string
				outcome    Outcome
				err        error
				branchName string
			}
			results := make([]buildResult, batchSize)

			var wg sync.WaitGroup
			for i, id := range batch {
				t, err := LoadTicket(ticketsDir, id)
				if err != nil {
					result.Stopped = "build_error"
					return result
				}

				if !config.Quiet {
					fmt.Printf("loop: building %s — %s (worker %d/%d)\n", id, t.Title, i+1, batchSize)
				}
				log.LoopTicketStart(id, t.Title)

				wg.Add(1)
				go func(idx int, ticketID string) {
					defer wg.Done()

					wtTicketsDir, branchName, err := createWorktree(ticketsDir, ticketID)
					if err != nil {
						results[idx] = buildResult{id: ticketID, err: fmt.Errorf("worktree: %w", err)}
						return
					}

					wtTicket, err := LoadTicket(wtTicketsDir, ticketID)
					if err != nil {
						results[idx] = buildResult{id: ticketID, err: err, branchName: branchName}
						return
					}

					outcome, buildErr := RunBuild(wtTicketsDir, wtTicket, p, log, config.Verbose)
					results[idx] = buildResult{
						id:         ticketID,
						outcome:    outcome,
						err:        buildErr,
						branchName: branchName,
					}
				}(i, id)
			}
			wg.Wait()

			// Sequential merge phase
			for _, br := range results {
				result.Processed++

				if br.err != nil {
					fmt.Fprintf(os.Stderr, "loop: build error for %s: %v\n", br.id, br.err)
					result.Failed++
					if br.branchName != "" {
						removeWorktree(ticketsDir, br.id, br.branchName)
					}
					log.LoopTicketComplete(br.id, "fail")
					continue
				}

				// Merge worktree branch back into main
				if br.branchName != "" {
					if err := mergeWorktree(ticketsDir, br.branchName); err != nil {
						fmt.Fprintf(os.Stderr, "loop: merge failed for %s: %v\n", br.id, err)
						removeWorktree(ticketsDir, br.id, br.branchName)
						result.Failed++
						log.LoopTicketComplete(br.id, "fail")
						continue
					}
					removeWorktree(ticketsDir, br.id, br.branchName)
				}

				switch br.outcome {
				case OutcomeSucceed:
					result.Succeeded++
				case OutcomeFail:
					result.Failed++
				case OutcomeBlocked:
					result.Blocked++
				case OutcomeDecompose:
					result.Decomposed++
				}

				outcomeStr := outcomeString(br.outcome)
				if !config.Quiet {
					fmt.Printf("loop: %s %s\n", br.id, strings.ToUpper(outcomeStr))
				}
				log.LoopTicketComplete(br.id, outcomeStr)
			}
		}
	}
}

// createWorktree creates a git worktree for a parallel ticket build.
// Returns the worktree's ticketsDir and branch name.
func createWorktree(mainTicketsDir, ticketID string) (string, string, error) {
	projectRoot := ProjectRoot(mainTicketsDir)
	branchName := "ko-worker-" + ticketID

	// Find the git toplevel so we can compute the relative path from git root
	// to the tickets dir. This handles projects in subdirectories (e.g. research/
	// inside discovery-zone).
	gitTop := exec.Command("git", "rev-parse", "--show-toplevel")
	gitTop.Dir = projectRoot
	topOut, err := gitTop.Output()
	if err != nil {
		return "", "", fmt.Errorf("git rev-parse --show-toplevel: %v", err)
	}
	gitRoot := strings.TrimSpace(string(topOut))

	absTicketsDir, err := filepath.Abs(mainTicketsDir)
	if err != nil {
		return "", "", err
	}
	relTicketsDir, err := filepath.Rel(gitRoot, absTicketsDir)
	if err != nil {
		return "", "", err
	}

	worktreeBase := filepath.Join(os.TempDir(), fmt.Sprintf("ko-workers-%d", os.Getpid()))
	worktreeRoot := filepath.Join(worktreeBase, ticketID)

	// Clean up stale branch from a prior crash if it exists
	delBranch := exec.Command("git", "branch", "-D", branchName)
	delBranch.Dir = projectRoot
	delBranch.CombinedOutput() // best-effort, ignore errors if branch doesn't exist

	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreeRoot, "HEAD")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("git worktree add: %v\n%s", err, string(out))
	}

	wtTicketsDir := filepath.Join(worktreeRoot, relTicketsDir)

	// Copy the ticket file into the worktree if it's uncommitted (untracked
	// or modified in the working tree). The worktree is created from HEAD,
	// so any ticket created since the last commit won't exist there.
	srcTicket := filepath.Join(absTicketsDir, ticketID+".md")
	dstTicket := filepath.Join(wtTicketsDir, ticketID+".md")
	if data, err := os.ReadFile(srcTicket); err == nil {
		os.MkdirAll(wtTicketsDir, 0755) // ensure dir exists
		os.WriteFile(dstTicket, data, 0644)
	}

	return wtTicketsDir, branchName, nil
}

// mergeWorktree merges a worktree branch back into the current branch.
func mergeWorktree(mainTicketsDir, branchName string) error {
	projectRoot := ProjectRoot(mainTicketsDir)
	cmd := exec.Command("git", "merge", branchName, "--no-edit")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge %s: %v\n%s", branchName, err, string(out))
	}
	return nil
}

// removeWorktree removes a worktree and deletes its temp branch.
func removeWorktree(mainTicketsDir, ticketID, branchName string) {
	projectRoot := ProjectRoot(mainTicketsDir)
	worktreeBase := filepath.Join(os.TempDir(), fmt.Sprintf("ko-workers-%d", os.Getpid()))
	worktreeRoot := filepath.Join(worktreeBase, ticketID)

	cmd := exec.Command("git", "worktree", "remove", "--force", worktreeRoot)
	cmd.Dir = projectRoot
	cmd.CombinedOutput() // best-effort

	cmd = exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = projectRoot
	cmd.CombinedOutput() // best-effort
}

// pruneWorktrees cleans up stale worktrees from prior crashes.
func pruneWorktrees(mainTicketsDir string) {
	projectRoot := ProjectRoot(mainTicketsDir)
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = projectRoot
	cmd.CombinedOutput() // best-effort
}

// runLoopHooks executes a list of shell commands with loop summary env vars set.
func runLoopHooks(ticketsDir string, hooks []string, result LoopResult, elapsed time.Duration) error {
	if len(hooks) == 0 {
		return nil
	}

	projectRoot := ProjectRoot(ticketsDir)

	for _, hook := range hooks {
		expanded := os.Expand(hook, func(key string) string {
			switch key {
			case "LOOP_PROCESSED":
				return strconv.Itoa(result.Processed)
			case "LOOP_SUCCEEDED":
				return strconv.Itoa(result.Succeeded)
			case "LOOP_FAILED":
				return strconv.Itoa(result.Failed)
			case "LOOP_BLOCKED":
				return strconv.Itoa(result.Blocked)
			case "LOOP_DECOMPOSED":
				return strconv.Itoa(result.Decomposed)
			case "LOOP_STOPPED":
				return result.Stopped
			case "LOOP_RUNTIME_SECONDS":
				return strconv.FormatFloat(elapsed.Seconds(), 'f', 2, 64)
			default:
				return os.Getenv(key)
			}
		})

		cmd := exec.Command("sh", "-c", expanded)
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(),
			"LOOP_PROCESSED="+strconv.Itoa(result.Processed),
			"LOOP_SUCCEEDED="+strconv.Itoa(result.Succeeded),
			"LOOP_FAILED="+strconv.Itoa(result.Failed),
			"LOOP_BLOCKED="+strconv.Itoa(result.Blocked),
			"LOOP_DECOMPOSED="+strconv.Itoa(result.Decomposed),
			"LOOP_STOPPED="+result.Stopped,
			"LOOP_RUNTIME_SECONDS="+strconv.FormatFloat(elapsed.Seconds(), 'f', 2, 64),
		)

		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("hook '%s' failed: %v\n%s", hook, err, string(out))
		}
	}
	return nil
}
