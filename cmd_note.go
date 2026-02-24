package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func cmdAddNote(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko note: ticket ID required")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}

	// Determine note content: stdin takes precedence over args if stdin is a pipe
	var note string
	stdinInfo, err := os.Stdin.Stat()
	isStdinPipe := err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0

	if isStdinPipe {
		// Stdin is a pipe (not a terminal), read from it
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko note: failed to read from stdin: %v\n", err)
			return 1
		}
		note = strings.TrimSpace(string(stdinBytes))
		// If stdin was empty and we have args, use args as fallback
		if note == "" && len(args) >= 2 {
			note = strings.Join(args[1:], " ")
		}
	} else {
		// Stdin is a terminal, require args
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "ko note: note text required (from args or stdin)")
			return 1
		}
		note = strings.Join(args[1:], " ")
	}

	// Final check: ensure we have note content
	if note == "" {
		fmt.Fprintln(os.Stderr, "ko note: note text cannot be empty")
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}

	AddNote(t, note)

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "note", nil)

	fmt.Printf("Note added to %s\n", id)
	return 0
}
