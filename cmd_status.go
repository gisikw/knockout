package main

import (
	"fmt"
	"os"
	"strings"
)

func cmdStatus(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko status: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko status: ticket ID and status required")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko status: %v\n", err)
		return 1
	}

	newStatus := args[1]
	if !ValidStatus(newStatus) {
		fmt.Fprintf(os.Stderr, "ko status: invalid status '%s'\nvalid statuses: %s\n", newStatus, strings.Join(Statuses, " "))
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko status: %v\n", err)
		return 1
	}

	oldStatus := t.Status
	t.Status = newStatus
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko status: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "status", map[string]interface{}{
		"from": oldStatus,
		"to":   newStatus,
	})

	fmt.Printf("%s -> %s\n", id, newStatus)
	return 0
}

func cmdStart(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko start: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "in_progress"))
}

func cmdClose(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko close: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "closed"))
}

func cmdOpen(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko open: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "open"))
}

func cmdBlock(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko block: ticket ID required")
		return 1
	}

	// Check if --questions flag is present
	hasQuestions := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--questions") {
			hasQuestions = true
			break
		}
	}

	// If --questions is present, pass through with --status=blocked
	if hasQuestions {
		return cmdUpdate(append(args, "--status=blocked"))
	}

	// If a reason is provided (args[1]), use it
	if len(args) > 1 {
		// Collect all remaining args as the reason
		reason := strings.Join(args[1:], " ")
		return cmdUpdate([]string{args[0], "--status=blocked", "-d", reason})
	}

	// No reason provided, just set status to blocked
	return cmdUpdate([]string{args[0], "--status=blocked"})
}

// ValidatePlanQuestions validates a slice of PlanQuestions.
// Returns an error if any required fields are missing or invalid.
func ValidatePlanQuestions(questions []PlanQuestion) error {
	for i, q := range questions {
		if q.ID == "" {
			return fmt.Errorf("question %d: missing required field 'id'", i)
		}
		if q.Question == "" {
			return fmt.Errorf("question %d (%s): missing required field 'question'", i, q.ID)
		}
		if len(q.Options) == 0 {
			return fmt.Errorf("question %d (%s): missing required field 'options' (must have at least one option)", i, q.ID)
		}
		for j, opt := range q.Options {
			if opt.Label == "" {
				return fmt.Errorf("question %d (%s), option %d: missing required field 'label'", i, q.ID, j)
			}
			if opt.Value == "" {
				return fmt.Errorf("question %d (%s), option %d: missing required field 'value'", i, q.ID, j)
			}
		}
	}
	return nil
}
