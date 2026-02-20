package main

import (
	"fmt"
	"os"
	"strings"
)

func cmdShow(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko show: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko show: ticket ID required")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko show: %v\n", err)
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko show: %v\n", err)
		return 1
	}

	// Print frontmatter
	fmt.Printf("id: %s\n", t.ID)
	fmt.Printf("status: %s\n", t.Status)
	fmt.Printf("type: %s\n", t.Type)
	fmt.Printf("priority: %d\n", t.Priority)
	fmt.Printf("deps: [%s]\n", strings.Join(t.Deps, ", "))
	fmt.Printf("links: [%s]\n", strings.Join(t.Links, ", "))
	fmt.Printf("created: %s\n", t.Created)
	if t.Assignee != "" {
		fmt.Printf("assignee: %s\n", t.Assignee)
	}
	if t.Parent != "" {
		fmt.Printf("parent: %s\n", t.Parent)
	}
	if t.ExternalRef != "" {
		fmt.Printf("external-ref: %s\n", t.ExternalRef)
	}
	if len(t.Tags) > 0 {
		fmt.Printf("tags: [%s]\n", strings.Join(t.Tags, ", "))
	}
	fmt.Println()
	fmt.Printf("# %s\n", t.Title)

	// Blockers section: deps that are not closed
	openDeps := openDeps(ticketsDir, t.Deps)
	if len(openDeps) > 0 {
		fmt.Println()
		fmt.Println("## Blockers")
		for _, d := range openDeps {
			fmt.Printf("  %s\n", d)
		}
	}

	// Blocking section: tickets that depend on this one
	blocking := findBlocking(ticketsDir, t.ID)
	if len(blocking) > 0 {
		fmt.Println()
		fmt.Println("## Blocking")
		for _, b := range blocking {
			fmt.Printf("  %s\n", b)
		}
	}

	// Children section
	children := findChildren(ticketsDir, t.ID)
	if len(children) > 0 {
		fmt.Println()
		fmt.Println("## Children")
		for _, c := range children {
			fmt.Printf("  %s\n", c)
		}
	}

	// Body content
	if t.Body != "" {
		fmt.Print(t.Body)
	}

	return 0
}

// openDeps returns dep IDs that are not in closed status.
func openDeps(ticketsDir string, deps []string) []string {
	var open []string
	for _, depID := range deps {
		t, err := LoadTicket(ticketsDir, depID)
		if err != nil {
			open = append(open, depID)
			continue
		}
		if t.Status != "closed" {
			open = append(open, depID)
		}
	}
	return open
}

// findBlocking returns IDs of tickets that have this ticket in their deps.
func findBlocking(ticketsDir, id string) []string {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return nil
	}
	var blocking []string
	for _, t := range tickets {
		for _, dep := range t.Deps {
			if dep == id {
				blocking = append(blocking, t.ID)
				break
			}
		}
	}
	return blocking
}

// findChildren returns IDs of tickets whose parent is this ticket.
func findChildren(ticketsDir, id string) []string {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return nil
	}
	var children []string
	for _, t := range tickets {
		if t.Parent == id {
			children = append(children, t.ID)
		}
	}
	return children
}
