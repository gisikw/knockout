package main

import (
	"fmt"
	"os"
)

func cmdLink(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko link: two ticket IDs required")
		return 1
	}

	id1, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	id2, err := ResolveID(ticketsDir, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	t1, err := LoadTicket(ticketsDir, id1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	t2, err := LoadTicket(ticketsDir, id2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	// Check if already linked
	for _, l := range t1.Links {
		if l == id2 {
			fmt.Printf("Link %s <-> %s already exists\n", id1, id2)
			return 0
		}
	}

	t1.Links = append(t1.Links, id2)
	t2.Links = append(t2.Links, id1)

	if err := SaveTicket(ticketsDir, t1); err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}
	if err := SaveTicket(ticketsDir, t2); err != nil {
		fmt.Fprintf(os.Stderr, "ko link: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id1, "link", map[string]interface{}{
		"linked": id2,
	})

	fmt.Printf("Linked %s <-> %s\n", id1, id2)
	return 0
}

func cmdUnlink(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko unlink: two ticket IDs required")
		return 1
	}

	id1, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	id2, err := ResolveID(ticketsDir, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	t1, err := LoadTicket(ticketsDir, id1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	t2, err := LoadTicket(ticketsDir, id2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	t1.Links = removeFromSlice(t1.Links, id2)
	t2.Links = removeFromSlice(t2.Links, id1)

	if err := SaveTicket(ticketsDir, t1); err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}
	if err := SaveTicket(ticketsDir, t2); err != nil {
		fmt.Fprintf(os.Stderr, "ko unlink: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id1, "unlink", map[string]interface{}{
		"unlinked": id2,
	})

	fmt.Printf("Unlinked %s <-> %s\n", id1, id2)
	return 0
}

func removeFromSlice(s []string, val string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}
