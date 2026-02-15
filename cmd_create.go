package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdCreate(args []string) int {
	args = reorderArgs(args, map[string]bool{
		"d": true, "t": true, "p": true, "a": true,
		"parent": true, "external-ref": true, "design": true,
		"acceptance": true, "tags": true,
	})

	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	desc := fs.String("d", "", "description")
	typ := fs.String("t", "", "ticket type")
	priority := fs.Int("p", -1, "priority (0-4)")
	assignee := fs.String("a", "", "assignee")
	parent := fs.String("parent", "", "parent ticket ID")
	extRef := fs.String("external-ref", "", "external reference")
	design := fs.String("design", "", "design notes")
	acceptance := fs.String("acceptance", "", "acceptance criteria")
	tags := fs.String("tags", "", "comma-separated tags")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko create: %v\n", err)
		return 1
	}

	title := "Untitled"
	if fs.NArg() > 0 {
		title = fs.Arg(0)
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		// Create .tickets in current directory
		ticketsDir = ".tickets"
	}
	if err := EnsureTicketsDir(ticketsDir); err != nil {
		fmt.Fprintf(os.Stderr, "ko create: %v\n", err)
		return 1
	}

	// Determine prefix from existing tickets or default to "ko"
	prefix := detectPrefix(ticketsDir)

	var t *Ticket
	if *parent != "" {
		// Resolve parent ID
		parentID, err := ResolveID(ticketsDir, *parent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko create: %v\n", err)
			return 1
		}
		t = NewChildTicket(parentID, title)
	} else {
		t = NewTicket(prefix, title)
	}

	if *desc != "" {
		t.Body += "\n" + *desc + "\n"
	}
	if *typ != "" {
		t.Type = *typ
	}
	if *priority >= 0 {
		t.Priority = *priority
	}
	if *assignee != "" {
		t.Assignee = *assignee
	}
	if *extRef != "" {
		t.ExternalRef = *extRef
	}
	if *design != "" {
		t.Body += "\n## Design\n\n" + *design + "\n"
	}
	if *acceptance != "" {
		t.Body += "\n## Acceptance Criteria\n\n" + *acceptance + "\n"
	}
	if *tags != "" {
		t.Tags = strings.Split(*tags, ",")
		for i, tag := range t.Tags {
			t.Tags[i] = strings.TrimSpace(tag)
		}
	}

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko create: %v\n", err)
		return 1
	}

	fmt.Println(t.ID)
	return 0
}

// detectPrefix looks at existing ticket files to infer the project prefix.
// Falls back to "ko".
func detectPrefix(ticketsDir string) string {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return "ko"
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		// Root ticket: prefix-hash (no dots)
		if !strings.Contains(id, ".") {
			if idx := strings.Index(id, "-"); idx > 0 {
				return id[:idx]
			}
		}
	}
	return "ko"
}
