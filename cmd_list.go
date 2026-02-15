package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdLs(args []string) int {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	statusFilter := fs.String("status", "", "filter by status")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	for _, t := range tickets {
		if *statusFilter != "" && t.Status != *statusFilter {
			continue
		}
		// Default: show non-closed tickets
		if *statusFilter == "" && t.Status == "closed" {
			continue
		}
		line := fmt.Sprintf("%s [%s] (p%d) %s", t.ID, t.Status, t.Priority, t.Title)
		if len(t.Deps) > 0 {
			line += fmt.Sprintf(" <- [%s]", strings.Join(t.Deps, ", "))
		}
		fmt.Println(line)
	}
	return 0
}

func cmdReady(args []string) int {
	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}

	var ready []*Ticket
	for _, t := range tickets {
		if IsReady(t.Status, AllDepsResolved(ticketsDir, t.Deps)) {
			ready = append(ready, t)
		}
	}

	SortByPriorityThenID(ready)

	for _, t := range ready {
		fmt.Printf("%s [%s] (p%d) %s\n", t.ID, t.Status, t.Priority, t.Title)
	}
	return 0
}

func cmdBlocked(args []string) int {
	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko blocked: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko blocked: %v\n", err)
		return 1
	}

	for _, t := range tickets {
		if t.Status == "closed" {
			continue
		}
		if len(t.Deps) == 0 {
			continue
		}
		if AllDepsResolved(ticketsDir, t.Deps) {
			continue
		}
		// Has unresolved deps
		fmt.Printf("%s [%s] (p%d) %s <- [%s]\n", t.ID, t.Status, t.Priority, t.Title, strings.Join(t.Deps, ", "))
	}
	return 0
}

func cmdClosed(args []string) int {
	fs := flag.NewFlagSet("closed", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "max tickets to show")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko closed: %v\n", err)
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko closed: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko closed: %v\n", err)
		return 1
	}

	count := 0
	for _, t := range tickets {
		if t.Status != "closed" {
			continue
		}
		fmt.Printf("%s [closed] (p%d) %s\n", t.ID, t.Priority, t.Title)
		count++
		if *limit > 0 && count >= *limit {
			break
		}
	}
	return 0
}
