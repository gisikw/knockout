package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func cmdCreate(args []string) int {
	if os.Getenv("KO_NO_CREATE") != "" {
		fmt.Fprintln(os.Stderr, "ko create: disabled — running in a loop context where creating new tickets could cause runaway expansion and incur significant costs")
		return 1
	}

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

	// Determine prefix from existing tickets or derive from directory name
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
// Falls back to deriving from the project root directory name.
func detectPrefix(ticketsDir string) string {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		absDir, absErr := filepath.Abs(ticketsDir)
		if absErr != nil {
			return DerivePrefix(filepath.Base(filepath.Dir(ticketsDir)))
		}
		return DerivePrefix(filepath.Base(filepath.Dir(absDir)))
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
	// No existing root tickets — derive from directory name.
	// Resolve to absolute path so relative paths like ".tickets" work correctly.
	absDir, err := filepath.Abs(ticketsDir)
	if err != nil {
		return DerivePrefix(filepath.Base(filepath.Dir(ticketsDir)))
	}
	return DerivePrefix(filepath.Base(filepath.Dir(absDir)))
}

// DerivePrefix generates a ticket prefix from a directory name.
// Multi-segment names (split on - and _) use the first letter of each segment.
// Single-segment names use the first 3 characters.
// Always returns a lowercase string of at least 2 characters.
func DerivePrefix(dirName string) string {
	dirName = strings.ToLower(dirName)

	// Split on hyphens and underscores
	segments := strings.FieldsFunc(dirName, func(r rune) bool {
		return r == '-' || r == '_'
	})

	if len(segments) > 1 {
		var b strings.Builder
		for _, seg := range segments {
			if len(seg) > 0 {
				b.WriteRune(rune(seg[0]))
			}
		}
		prefix := b.String()
		if len(prefix) >= 2 {
			return prefix
		}
	}

	// Single segment or initials too short: use first 3 chars
	// Strip non-alphanumeric
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, dirName)

	if len(cleaned) >= 3 {
		return cleaned[:3]
	}
	if len(cleaned) >= 2 {
		return cleaned
	}
	return "ko"
}
