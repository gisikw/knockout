package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdClear(args []string) int {
	force := false
	for _, a := range args {
		if a == "--force" || a == "-f" {
			force = true
		}
	}

	if !force {
		fmt.Fprintln(os.Stderr, "ko clear: requires --force to remove all local tickets")
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko clear: %v\n", err)
		return 1
	}

	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko clear: %v\n", err)
		return 1
	}

	removed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(ticketsDir, e.Name())
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "ko clear: %v\n", err)
			return 1
		}
		removed++
	}

	fmt.Printf("removed %d tickets\n", removed)
	return 0
}
