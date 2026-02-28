package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func cmdCreate(args []string) int {
	if os.Getenv("KO_NO_CREATE") != "" {
		fmt.Fprintln(os.Stderr, "ko add: disabled — running in a loop context where creating new tickets could cause runaway expansion and incur significant costs")
		return 1
	}

	args = reorderArgs(args, map[string]bool{
		"d": true, "t": true, "p": true, "a": true,
		"parent": true, "external-ref": true, "design": true,
		"acceptance": true, "tags": true, "project": true,
		"snooze": true, "triage": true,
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
	projectTag := fs.String("project", "", "target project tag")
	snooze := fs.String("snooze", "", "snooze date (ISO 8601, e.g. 2026-05-01)")
	triage := fs.String("triage", "", "triage note (free text)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	title := "Untitled"
	if fs.NArg() > 0 {
		title = fs.Arg(0)
	}

	// Determine description source: stdin > second positional arg > -d flag
	// Only read stdin if: (a) no -d flag provided, (b) no second positional arg,
	// and (c) stdin is actually a pipe with data (not just an inherited pipe from
	// a parent process). We use a 0-byte read with a deadline to avoid blocking
	// forever on pipes that never send data.
	var descFromInput string
	if *desc == "" && fs.NArg() <= 1 {
		stdinInfo, err := os.Stdin.Stat()
		isStdinPipe := err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0
		if isStdinPipe {
			// Set a short read deadline to avoid blocking on empty pipes.
			// os.Stdin doesn't support deadlines, so use a non-blocking peek
			// via a goroutine with a timeout.
			done := make(chan []byte, 1)
			go func() {
				b, _ := io.ReadAll(os.Stdin)
				done <- b
			}()
			select {
			case stdinBytes := <-done:
				descFromInput = strings.TrimSpace(string(stdinBytes))
			case <-time.After(50 * time.Millisecond):
				// No data on stdin within 50ms — treat as no input
			}
		}
	}

	// If stdin is empty or not a pipe, check for second positional arg
	if descFromInput == "" && fs.NArg() > 1 {
		descFromInput = fs.Arg(1)
	}

	// Find local project context
	localRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	// Determine target project from --project flag
	var targetPath string
	if *projectTag != "" {
		// Route to specified project
		regPath := RegistryPath()
		if regPath == "" {
			fmt.Fprintf(os.Stderr, "ko add: cannot determine config directory for project routing\n")
			return 1
		}
		reg, err := LoadRegistry(regPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
			return 1
		}
		projectPath, ok := reg.Projects[*projectTag]
		if !ok {
			fmt.Fprintf(os.Stderr, "ko add: unknown project '%s'\n", *projectTag)
			return 1
		}
		targetPath = projectPath
	} else {
		// No --project flag: create ticket locally
		targetPath = localRoot
	}

	// Ensure target tickets directory exists
	ticketsDir := resolveTicketsDir(targetPath)
	if err := EnsureTicketsDir(ticketsDir); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	// Determine prefix from existing tickets or derive from directory name
	prefix := detectPrefix(ticketsDir)

	var t *Ticket
	if *parent != "" {
		// Resolve parent ID (cross-project prefix lookup supported)
		_, parentID, err := ResolveTicket(ticketsDir, *parent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
			return 1
		}
		t = NewChildTicket(parentID, title)
	} else {
		t = NewTicket(prefix, title)
	}

	t.Status = "open"

	// Apply strict priority for description: stdin > arg > -d flag
	if descFromInput != "" {
		t.Body += "\n" + descFromInput + "\n"
	} else if *desc != "" {
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

	// Apply explicit tags from --tags flag
	if *tags != "" {
		var ticketTags []string
		for _, tag := range strings.Split(*tags, ",") {
			ticketTags = append(ticketTags, strings.TrimSpace(tag))
		}
		t.Tags = ticketTags
	}

	if *snooze != "" {
		if _, err := time.Parse("2006-01-02", *snooze); err != nil {
			fmt.Fprintf(os.Stderr, "ko add: invalid snooze date %q: must be ISO 8601 format (e.g. 2026-05-01)\n", *snooze)
			return 1
		}
		t.Snooze = *snooze
	}
	if *triage != "" {
		t.Triage = *triage
	}

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, t.ID, "create", map[string]interface{}{
		"title": t.Title,
	})

	if *triage != "" {
		maybeAutoTriage(ticketsDir, t.ID)
	}

	fmt.Println(t.ID)
	return 0
}

// ReadPrefix reads the persisted prefix from .ko/config.yaml (project.prefix)
// or falls back to .ko/prefix file for backwards compatibility.
// Returns "" if neither exists.
func ReadPrefix(ticketsDir string) string {
	root := ProjectRoot(ticketsDir)

	// Try reading from unified config.yaml first
	configPath := filepath.Join(root, ".ko", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		if config, err := LoadConfig(configPath); err == nil && config.Project.Prefix != "" {
			return config.Project.Prefix
		}
	}

	// Fall back to legacy .ko/prefix file
	data, err := os.ReadFile(filepath.Join(root, ".ko", "prefix"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// WritePrefix persists the prefix to .ko/prefix (legacy format).
// Deprecated: New code should write to .ko/config.yaml instead.
// Kept for backwards compatibility with older pipelines.
func WritePrefix(ticketsDir, prefix string) error {
	root := ProjectRoot(ticketsDir)
	koDir := filepath.Join(root, ".ko")
	if err := os.MkdirAll(koDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(koDir, "prefix"), []byte(prefix+"\n"), 0644)
}

// WriteConfigPrefix writes the prefix to .ko/config.yaml.
// If config.yaml already exists, it updates the project.prefix field.
// If it doesn't exist, it creates a minimal config with just the project section.
func WriteConfigPrefix(ticketsDir, prefix string) error {
	root := ProjectRoot(ticketsDir)
	koDir := filepath.Join(root, ".ko")
	if err := os.MkdirAll(koDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(koDir, "config.yaml")

	// Check if config.yaml already exists
	existingData, err := os.ReadFile(configPath)
	if err == nil {
		// Config exists - update the project.prefix field
		content := string(existingData)
		lines := strings.Split(content, "\n")
		var result []string
		inProject := false
		foundPrefix := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Detect project: section
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				if trimmed == "project:" {
					inProject = true
					result = append(result, line)
					continue
				} else if strings.HasSuffix(trimmed, ":") {
					inProject = false
				}
			}

			// If in project section and this is the prefix line, replace it
			if inProject {
				if key, _, ok := parseYAMLLine(trimmed); ok && key == "prefix" {
					result = append(result, "  prefix: "+prefix)
					foundPrefix = true
					continue
				}
			}

			result = append(result, line)
		}

		// If we didn't find a prefix line but are in project section, we need to add it
		// For now, just rewrite if prefix wasn't found
		if !foundPrefix {
			// Append to end of project section or create project section
			return os.WriteFile(configPath, []byte(strings.Join(result, "\n")), 0644)
		}

		return os.WriteFile(configPath, []byte(strings.Join(result, "\n")), 0644)
	}

	// Config doesn't exist - create minimal config with just project.prefix
	minimalConfig := fmt.Sprintf("project:\n  prefix: %s\n", prefix)
	return os.WriteFile(configPath, []byte(minimalConfig), 0644)
}

// detectPrefix looks at existing ticket files to infer the project prefix.
// Falls back to deriving from the project root directory name.
// Persists the result to .ko/prefix for future stability.
func detectPrefix(ticketsDir string) string {
	// Check persisted prefix first
	if p := ReadPrefix(ticketsDir); p != "" {
		return p
	}

	// Scan existing tickets
	entries, err := os.ReadDir(ticketsDir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}
			id := strings.TrimSuffix(name, ".md")
			// Root ticket: prefix-hash (no dots)
			if !strings.Contains(id, ".") {
				if idx := strings.Index(id, "-"); idx > 0 {
					prefix := id[:idx]
					WritePrefix(ticketsDir, prefix)
					return prefix
				}
			}
		}
	}

	// No existing root tickets — derive from project root directory name.
	root := ProjectRoot(ticketsDir)
	prefix := DerivePrefix(filepath.Base(root))
	WritePrefix(ticketsDir, prefix)
	return prefix
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
