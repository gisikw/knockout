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

func cmdReopen(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko reopen: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "open"))
}

func cmdBlock(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko block: ticket ID required")
		return 1
	}
	return cmdStatus(append(args, "blocked"))
}
