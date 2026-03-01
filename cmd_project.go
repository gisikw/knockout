package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type projectJSON struct {
	Tag       string `json:"tag"`
	Path      string `json:"path"`
	IsDefault bool   `json:"is_default"`
}

func cmdProject(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ko project: subcommand required (set, ls)")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  ko project set #<tag> [--prefix=p] [--default]")
		fmt.Fprintln(os.Stderr, "  ko project ls")
		return 1
	}

	subcmd := args[0]
	rest := args[1:]

	switch subcmd {
	case "set":
		return cmdProjectSet(rest)
	case "ls":
		return cmdProjectLs(rest)
	default:
		fmt.Fprintf(os.Stderr, "ko project: unknown subcommand '%s'\n", subcmd)
		fmt.Fprintln(os.Stderr, "Valid subcommands: set, ls")
		return 1
	}
}

func cmdProjectSet(args []string) int {
	// Parse flags
	var prefix string
	var setDefault bool
	var tag string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--prefix=") {
			prefix = strings.TrimPrefix(arg, "--prefix=")
		} else if arg == "--default" {
			setDefault = true
		} else if strings.HasPrefix(arg, "#") {
			tag = CleanTag(arg)
		} else {
			fmt.Fprintf(os.Stderr, "ko project set: unexpected argument '%s'\n", arg)
			return 1
		}
	}

	// Validate tag is provided
	if tag == "" {
		fmt.Fprintln(os.Stderr, "ko project set: #tag argument required")
		fmt.Fprintln(os.Stderr, "Usage: ko project set #<tag> [--prefix=p] [--default]")
		return 1
	}

	// Validate prefix if provided
	if prefix != "" && len(prefix) < 2 {
		fmt.Fprintln(os.Stderr, "ko project set: prefix must be at least 2 characters")
		return 1
	}

	// Get current working directory as project root
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko project set: %v\n", err)
		return 1
	}

	ticketsDir := filepath.Join(root, ".ko", "tickets")

	// Create .ko/tickets directory if it doesn't exist (init)
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ko project set: %v\n", err)
		return 1
	}

	// Write prefix to config.yaml if provided
	if prefix != "" {
		if err := WriteConfigPrefix(ticketsDir, prefix); err != nil {
			fmt.Fprintf(os.Stderr, "ko project set: %v\n", err)
			return 1
		}
	}

	// Register in global registry
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko project set: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko project set: %v\n", err)
		return 1
	}

	// Evict any existing entry for the same path under a different tag.
	for existingTag, existingPath := range reg.Projects {
		if existingPath == root && existingTag != tag {
			delete(reg.Projects, existingTag)
			delete(reg.Prefixes, existingTag)
			if reg.Default == existingTag {
				reg.Default = tag
			}
		}
	}

	reg.Projects[tag] = root

	// Detect and store ticket prefix if not explicitly provided
	if prefix == "" {
		if detected := detectPrefixFromDir(ticketsDir); detected != "" {
			reg.Prefixes[tag] = detected
		}
	} else {
		reg.Prefixes[tag] = prefix
	}

	// Set as default if requested
	if setDefault {
		reg.Default = tag
	}

	if err := SaveRegistry(regPath, reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko project set: %v\n", err)
		return 1
	}

	// Print success message
	if prefix != "" && setDefault {
		fmt.Printf("project %s initialized with prefix %q, registered, and set as default\n", tag, prefix)
	} else if prefix != "" {
		fmt.Printf("project %s initialized with prefix %q and registered\n", tag, prefix)
	} else if setDefault {
		fmt.Printf("project %s registered and set as default\n", tag)
	} else {
		fmt.Printf("project %s registered\n", tag)
	}

	return 0
}

func cmdProjectLs(args []string) int {
	fs := flag.NewFlagSet("project ls", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko project ls: %v\n", err)
		return 1
	}

	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko project ls: cannot determine config directory")
		return 1
	}

	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko project ls: %v\n", err)
		return 1
	}

	if len(reg.Projects) == 0 {
		if *jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("no projects registered")
		}
		return 0
	}

	// Sort tags
	keys := make([]string, 0, len(reg.Projects))
	for k := range reg.Projects {
		keys = append(keys, k)
	}
	sortStrings(keys)

	if *jsonOutput {
		projects := make([]projectJSON, 0, len(keys))
		for _, k := range keys {
			projects = append(projects, projectJSON{
				Tag:       k,
				Path:      reg.Projects[k],
				IsDefault: k == reg.Default,
			})
		}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(projects); err != nil {
			fmt.Fprintf(os.Stderr, "ko project ls: failed to encode JSON: %v\n", err)
			return 1
		}
	} else {
		// Print with default marker
		for _, k := range keys {
			marker := "  "
			if k == reg.Default {
				marker = "* "
			}
			fmt.Printf("%s%s\t%s\n", marker, k, reg.Projects[k])
		}
	}

	return 0
}
