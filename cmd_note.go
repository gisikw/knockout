package main

import (
	"fmt"
	"os"
	"strings"
)

func cmdAddNote(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko add-note: ticket ID and note text required")
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add-note: %v\n", err)
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add-note: %v\n", err)
		return 1
	}

	note := strings.Join(args[1:], " ")

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add-note: %v\n", err)
		return 1
	}

	AddNote(t, note)

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko add-note: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "note", nil)

	fmt.Printf("Note added to %s\n", id)
	return 0
}
