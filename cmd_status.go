package main

import (
	"encoding/json"
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

func cmdReopen(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko reopen: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "open"))
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

func cmdBlock(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko block: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko block: ticket ID required")
		return 1
	}

	// Parse flags
	var questionsJSON string
	var ticketID string
	for i := 0; i < len(args); i++ {
		if args[i] == "--questions" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "ko block: --questions flag requires a JSON argument")
				return 1
			}
			questionsJSON = args[i+1]
			i++ // skip next arg
		} else if ticketID == "" {
			ticketID = args[i]
		}
	}

	if ticketID == "" {
		fmt.Fprintln(os.Stderr, "ko block: ticket ID required")
		return 1
	}

	id, err := ResolveID(ticketsDir, ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko block: %v\n", err)
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko block: %v\n", err)
		return 1
	}

	// Parse and validate questions if provided
	if questionsJSON != "" {
		var questions []PlanQuestion
		if err := json.Unmarshal([]byte(questionsJSON), &questions); err != nil {
			fmt.Fprintf(os.Stderr, "ko block: invalid JSON: %v\n", err)
			return 1
		}

		if err := ValidatePlanQuestions(questions); err != nil {
			fmt.Fprintf(os.Stderr, "ko block: invalid questions: %v\n", err)
			return 1
		}

		t.PlanQuestions = questions
	}

	oldStatus := t.Status
	t.Status = "blocked"

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko block: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "status", map[string]interface{}{
		"from": oldStatus,
		"to":   "blocked",
	})

	fmt.Printf("%s -> blocked\n", id)
	return 0
}
