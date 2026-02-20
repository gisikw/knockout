package main

import (
	"fmt"
	"os"
)

func cmdDep(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko dep: subcommand or ticket IDs required")
		fmt.Fprintln(os.Stderr, "usage: ko dep <ticket> <dep>  |  ko dep tree <ticket>")
		return 1
	}

	// Check for tree subcommand before resolving project tag
	// (tree args may also contain a #project tag)
	if args[0] == "tree" {
		return cmdDepTree(args[1:])
	}

	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko dep: two ticket IDs required")
		return 1
	}

	ticketID, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}

	depID, err := ResolveID(ticketsDir, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}

	if ticketID == depID {
		fmt.Fprintf(os.Stderr, "ko dep: ticket cannot depend on itself\n")
		return 1
	}

	t, err := LoadTicket(ticketsDir, ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}

	// Check if dep already exists
	for _, d := range t.Deps {
		if d == depID {
			fmt.Printf("Dependency %s -> %s already exists\n", ticketID, depID)
			return 0
		}
	}

	t.Deps = append(t.Deps, depID)
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, ticketID, "dep", map[string]interface{}{
		"dep": depID,
	})

	fmt.Printf("Added dependency: %s -> %s\n", ticketID, depID)
	return 0
}

func cmdUndep(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko undep: two ticket IDs required")
		return 1
	}

	ticketID, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}

	depID, err := ResolveID(ticketsDir, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}

	t, err := LoadTicket(ticketsDir, ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}

	found := false
	newDeps := make([]string, 0, len(t.Deps))
	for _, d := range t.Deps {
		if d == depID {
			found = true
		} else {
			newDeps = append(newDeps, d)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "ko undep: dependency %s -> %s not found\n", ticketID, depID)
		return 1
	}

	t.Deps = newDeps
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, ticketID, "undep", map[string]interface{}{
		"dep": depID,
	})

	fmt.Printf("Removed dependency: %s -> %s\n", ticketID, depID)
	return 0
}

func cmdDepTree(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep tree: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko dep tree: ticket ID required")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep tree: %v\n", err)
		return 1
	}

	visited := make(map[string]bool)
	printDepTree(ticketsDir, id, "", visited)
	return 0
}

func printDepTree(ticketsDir, id, indent string, visited map[string]bool) {
	if visited[id] {
		return
	}
	visited[id] = true

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Printf("%s%s (not found)\n", indent, id)
		return
	}

	fmt.Printf("%s%s [%s] %s\n", indent, t.ID, t.Status, t.Title)
	for _, dep := range t.Deps {
		printDepTree(ticketsDir, dep, indent+"  ", visited)
	}
}
