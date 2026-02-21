package main

import (
	"fmt"
	"os"
	"strings"
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

// ReadyQueue returns the IDs of tickets ready to build, sorted by priority.
func ReadyQueue(ticketsDir string) ([]string, error) {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return nil, err
	}

	var ready []*Ticket
	for _, t := range tickets {
		if IsReady(t.Status, AllDepsResolved(ticketsDir, t.Deps)) {
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

// RunLoop burns down the ready queue, building one ticket at a time.
// Sets KO_NO_CREATE to prevent spawned agents from creating tickets.
// If stop is non-nil, the loop checks it between builds and exits with
// "signal" when closed.
func RunLoop(ticketsDir string, p *Pipeline, config LoopConfig, log *EventLogger, stop <-chan struct{}) LoopResult {
	os.Setenv("KO_NO_CREATE", "1")
	defer os.Unsetenv("KO_NO_CREATE")

	start := time.Now()
	result := LoopResult{}

	for {
		// Check for stop signal
		if stop != nil {
			select {
			case <-stop:
				result.Stopped = "signal"
				return result
			default:
			}
		}

		// Check limits
		if ok, reason := config.ShouldContinue(result.Processed, time.Since(start)); !ok {
			result.Stopped = reason
			return result
		}

		// Get next ready ticket
		queue, err := ReadyQueue(ticketsDir)
		if err != nil {
			result.Stopped = "build_error"
			return result
		}
		if len(queue) == 0 {
			result.Stopped = "empty"
			return result
		}

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
	}
}
