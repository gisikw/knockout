package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
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

	// Determine note content: positional args > stdin pipe > error
	var note string
	if len(args) >= 2 {
		// Positional args provided — use them, skip stdin
		note = strings.Join(args[1:], " ")
	} else {
		// No positional note text — try stdin if it's a pipe
		stdinInfo, err := os.Stdin.Stat()
		isStdinPipe := err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0
		if isStdinPipe {
			done := make(chan []byte, 1)
			go func() {
				b, _ := io.ReadAll(os.Stdin)
				done <- b
			}()
			select {
			case stdinBytes := <-done:
				note = strings.TrimSpace(string(stdinBytes))
			case <-time.After(50 * time.Millisecond):
				// No data on stdin within 50ms — treat as no input
			}
		}
		if note == "" {
			fmt.Fprintln(os.Stderr, "ko note: note text required (from args or stdin)")
			return 1
		}
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
