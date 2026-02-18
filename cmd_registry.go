package main

import (
	"fmt"
	"os"
)

func cmdRegister(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ko register #<tag>")
		return 1
	}

	tag := CleanTag(args[0])
	if tag == "" {
		fmt.Fprintln(os.Stderr, "usage: ko register #<tag>")
		return 1
	}

	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko register: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	// Determine project root from .tickets directory or cwd
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	reg.Projects[tag] = projectRoot

	// Detect and store ticket prefix
	ticketsDir := resolveTicketsDir(projectRoot)
	if prefix := detectPrefixFromDir(ticketsDir); prefix != "" {
		reg.Prefixes[tag] = prefix
	}

	if err := SaveRegistry(regPath, reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko register: %v\n", err)
		return 1
	}

	fmt.Printf("registered %s -> %s\n", tag, projectRoot)
	return 0
}

func cmdDefault(args []string) int {
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko default: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko default: %v\n", err)
		return 1
	}

	// No args: show current default
	if len(args) == 0 {
		if reg.Default == "" {
			fmt.Println("no default project set")
		} else {
			fmt.Println(reg.Default)
		}
		return 0
	}

	tag := CleanTag(args[0])

	// Verify tag is registered
	if _, ok := reg.Projects[tag]; !ok {
		fmt.Fprintf(os.Stderr, "ko default: '%s' is not registered\n", tag)
		return 1
	}

	reg.Default = tag

	if err := SaveRegistry(regPath, reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko default: %v\n", err)
		return 1
	}

	fmt.Printf("default: %s\n", tag)
	return 0
}

func cmdProjects(args []string) int {
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko projects: cannot determine config directory")
		return 1
	}

	if _, err := os.Stat(regPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ko projects: no project registry found (%s)\n", regPath)
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko projects: %v\n", err)
		return 1
	}

	if len(reg.Projects) == 0 {
		fmt.Println("no projects registered")
		return 0
	}

	keys := make([]string, 0, len(reg.Projects))
	for k := range reg.Projects {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		marker := "  "
		if k == reg.Default {
			marker = "* "
		}
		fmt.Printf("%s%s\t%s\n", marker, k, reg.Projects[k])
	}
	return 0
}

func cmdAdd(args []string) int {
	if os.Getenv("KO_NO_CREATE") != "" {
		fmt.Fprintln(os.Stderr, "ko add: disabled — running in a loop context where creating new tickets could cause runaway expansion and incur significant costs")
		return 1
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ko add '<title> [#tag ...]'")
		return 1
	}

	title := args[0]

	// Find local project context
	localRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	// Load registry (non-fatal if missing — just route locally)
	regPath := RegistryPath()
	reg := &Registry{Projects: map[string]string{}, Prefixes: map[string]string{}}
	if regPath != "" {
		loaded, loadErr := LoadRegistry(regPath)
		if loadErr == nil {
			reg = loaded
		}
	}

	decision := RouteTicket(title, reg, localRoot)

	// Ensure target tickets directory exists
	targetTicketsDir := resolveTicketsDir(decision.TargetPath)
	if err := EnsureTicketsDir(targetTicketsDir); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	// Create ticket in target project
	prefix := detectPrefix(targetTicketsDir)
	t := NewTicket(prefix, decision.Title)
	t.Status = decision.Status

	// Attach tags: routing tag (if unrecognized/captured) + extra tags
	var ticketTags []string
	if decision.IsCaptured && decision.RoutingTag != "" {
		ticketTags = append(ticketTags, decision.RoutingTag)
	}
	ticketTags = append(ticketTags, decision.ExtraTags...)
	if len(ticketTags) > 0 {
		t.Tags = ticketTags
	}

	if err := SaveTicket(targetTicketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}

	// If routed to a different project, create a closed audit ticket locally
	if decision.IsRouted {
		localTicketsDir := resolveTicketsDir(localRoot)
		if err := EnsureTicketsDir(localTicketsDir); err != nil {
			fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
			return 1
		}
		localPrefix := detectPrefix(localTicketsDir)
		audit := NewTicket(localPrefix, decision.Title)
		audit.Status = "closed"
		AddNote(audit, fmt.Sprintf("routed to #%s as %s", decision.RoutingTag, t.ID))
		if err := SaveTicket(localTicketsDir, audit); err != nil {
			fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
			return 1
		}
		fmt.Printf("%s -> #%s (%s)\n", audit.ID, decision.RoutingTag, t.ID)
	} else {
		fmt.Println(t.ID)
	}

	return 0
}

// findProjectRoot returns the absolute path to the project root.
// If a tickets directory is found, derives root via ProjectRoot.
// Otherwise, returns the current working directory.
func findProjectRoot() (string, error) {
	ticketsDir, err := FindTicketsDir()
	if err == nil {
		return ProjectRoot(ticketsDir), nil
	}
	return os.Getwd()
}
