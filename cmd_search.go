package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	projectFlag := fs.String("project", "", "Filter by project tag")
	statusFlag := fs.String("status", "", "Filter by status")
	typeFlag := fs.String("type", "", "Filter by ticket type")
	tagFlag := fs.String("tag", "", "Filter by ticket tag")
	limitFlag := fs.Int("limit", 50, "Maximum results")
	jsonFlag := fs.Bool("json", false, "Output as JSON")

	// Reorder args to handle flags after query
	reordered := reorderArgs(args, map[string]bool{
		"project": true,
		"status":  true,
		"type":    true,
		"tag":     true,
		"limit":   true,
	})

	if err := fs.Parse(reordered); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ko search <query> [--project=tag] [--status=X] [--type=X] [--tag=X] [--limit=50] [--json]")
		return 1
	}

	query := strings.Join(fs.Args(), " ")

	db, err := OpenReadDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko:", err)
		return 1
	}
	defer db.Close()

	results, err := db.SearchTickets(query, *projectFlag, *statusFlag, *typeFlag, *tagFlag, *limitFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: search:", err)
		return 1
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return 0
	}

	if *jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(results)
		return 0
	}

	// Text output
	for _, r := range results {
		fmt.Printf("%s [%s] [%s] (p%d) %s\n", r.TicketID, r.Project, r.Status, r.Priority, r.Title)
		if r.Snippet != "" {
			fmt.Printf("  %s\n", r.Snippet)
		}
	}

	if len(results) == *limitFlag {
		fmt.Printf("\n(showing first %d results, use --limit to see more)\n", *limitFlag)
	}

	return 0
}
