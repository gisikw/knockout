package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdStats(args []string) int {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	projectFlag := fs.String("project", "", "Filter by project tag")
	jsonFlag := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	db, err := OpenReadDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko:", err)
		return 1
	}
	defer db.Close()

	stats, err := db.QueryStats(*projectFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: stats:", err)
		return 1
	}

	if *jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(stats)
		return 0
	}

	// Text output
	fmt.Printf("Tickets: %d total\n", stats.Total)

	// By status
	if len(stats.ByStatus) > 0 {
		var parts []string
		for _, s := range stats.ByStatus {
			parts = append(parts, fmt.Sprintf("%s: %d", s.Status, s.Count))
		}
		fmt.Printf("  %s\n", strings.Join(parts, "  "))
	}

	fmt.Println()

	// By type
	if len(stats.ByType) > 0 {
		var parts []string
		for _, t := range stats.ByType {
			parts = append(parts, fmt.Sprintf("%s: %d", t.Type, t.Count))
		}
		fmt.Printf("By type:     %s\n", strings.Join(parts, "  "))
	}

	// By priority
	if len(stats.ByPriority) > 0 {
		var parts []string
		for _, p := range stats.ByPriority {
			parts = append(parts, fmt.Sprintf("p%d: %d", p.Priority, p.Count))
		}
		fmt.Printf("By priority: %s\n", strings.Join(parts, "  "))
	}

	fmt.Println()

	// Time-based stats
	fmt.Printf("This week:   +%d created, +%d closed\n", stats.CreatedThisWeek, stats.ClosedThisWeek)
	fmt.Printf("This month:  +%d created, +%d closed\n", stats.CreatedThisMonth, stats.ClosedThisMonth)

	fmt.Println()

	// Queue stats
	fmt.Printf("Ready: %d  Blocked: %d\n", stats.Ready, stats.Blocked)

	fmt.Println()

	// Build stats
	if stats.TotalBuilds > 0 {
		pct := float64(stats.Succeeded) / float64(stats.TotalBuilds) * 100
		fmt.Printf("Builds: %d total, %d succeeded (%.1f%%), %d failed\n",
			stats.TotalBuilds, stats.Succeeded, pct, stats.Failed)
	} else {
		fmt.Println("Builds: 0")
	}

	// Per-project breakdown
	if len(stats.ByProject) > 0 {
		fmt.Println()
		fmt.Println("By project:")
		for _, p := range stats.ByProject {
			fmt.Printf("  %-15s %d open / %d closed\n", p.Tag, p.Open, p.Closed)
		}
	}

	return 0
}
