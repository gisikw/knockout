package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func cmdInit(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ko init: prefix argument required")
		fmt.Fprintln(os.Stderr, "Usage: ko init <prefix>")
		return 1
	}
	prefix := args[0]
	if len(prefix) < 2 {
		fmt.Fprintln(os.Stderr, "ko init: prefix must be at least 2 characters")
		return 1
	}

	// Use cwd as the project root
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko init: %v\n", err)
		return 1
	}

	ticketsDir := filepath.Join(root, ".ko", "tickets")

	// Check if already initialized (check both config.yaml and legacy prefix file)
	configPath := filepath.Join(root, ".ko", "config.yaml")
	legacyPrefixPath := filepath.Join(root, ".ko", "prefix")
	if _, err := os.Stat(configPath); err == nil {
		existing := ReadPrefix(ticketsDir)
		fmt.Fprintf(os.Stderr, "ko init: already initialized with prefix %q (in config.yaml)\n", existing)
		return 1
	}
	if _, err := os.Stat(legacyPrefixPath); err == nil {
		existing := ReadPrefix(ticketsDir)
		fmt.Fprintf(os.Stderr, "ko init: already initialized with prefix %q\n", existing)
		return 1
	}

	// Create directories
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ko init: %v\n", err)
		return 1
	}

	// Write prefix to config.yaml (new unified format)
	if err := WriteConfigPrefix(ticketsDir, prefix); err != nil {
		fmt.Fprintf(os.Stderr, "ko init: %v\n", err)
		return 1
	}

	fmt.Printf("initialized %s with prefix %q\n", root, prefix)
	return 0
}
