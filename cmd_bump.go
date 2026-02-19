package main

import (
	"fmt"
	"os"
	"time"
)

func cmdBump(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ko bump <id>")
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko bump: %v\n", err)
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko bump: %v\n", err)
		return 1
	}

	path := TicketPath(ticketsDir, id)
	now := time.Now()
	if err := os.Chtimes(path, now, now); err != nil {
		fmt.Fprintf(os.Stderr, "ko bump: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "bump", nil)

	fmt.Println(id)
	return 0
}
