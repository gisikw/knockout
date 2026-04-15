package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// HistoryOutput is the JSON structure for ko history output.
type HistoryOutput struct {
	TicketID   string           `json:"ticket_id,omitempty"`
	Title      string           `json:"title,omitempty"`
	Builds     []BuildEntry     `json:"builds,omitempty"`
	Mutations  []MutationEntry  `json:"mutations,omitempty"`
}

func cmdHistory(args []string) int {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	projectFlag := fs.String("project", "", "Filter by project tag")
	limitFlag := fs.Int("limit", 20, "Maximum results per section")
	jsonFlag := fs.Bool("json", false, "Output as JSON")

	// Reorder args to handle flags after ticket ID
	reordered := reorderArgs(args, map[string]bool{
		"project": true,
		"limit":   true,
	})

	if err := fs.Parse(reordered); err != nil {
		return 1
	}

	db, err := OpenReadDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko:", err)
		return 1
	}
	defer db.Close()

	// Per-ticket mode if argument provided
	if fs.NArg() > 0 {
		ticketID := fs.Arg(0)
		return cmdHistoryTicket(db, ticketID, *limitFlag, *jsonFlag)
	}

	// Global mode
	return cmdHistoryGlobal(db, *projectFlag, *limitFlag, *jsonFlag)
}

func cmdHistoryTicket(db *DB, ticketID string, limit int, asJSON bool) int {
	title, err := db.GetTicketTitle(ticketID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko:", err)
		return 1
	}

	builds, err := db.QueryTicketBuilds(ticketID, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: history:", err)
		return 1
	}

	mutations, err := db.QueryTicketMutations(ticketID, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: history:", err)
		return 1
	}

	if asJSON {
		out := HistoryOutput{
			TicketID:  ticketID,
			Title:     title,
			Builds:    builds,
			Mutations: mutations,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return 0
	}

	// Text output
	fmt.Printf("%s: %s\n\n", ticketID, title)

	if len(builds) > 0 {
		fmt.Println("Builds:")
		for _, b := range builds {
			outcome := "in progress"
			if b.Outcome.Valid {
				outcome = b.Outcome.String
			}
			dur := ""
			if b.Duration != "" {
				dur = fmt.Sprintf(" (%s)", b.Duration)
			}
			fmt.Printf("  %s  %-8s %s%s\n", formatTime(b.StartedAt), outcome, b.Workflow, dur)
		}
		fmt.Println()
	}

	if len(mutations) > 0 {
		fmt.Println("Events:")
		for _, m := range mutations {
			payload := ""
			if m.Payload.Valid && m.Payload.String != "" && m.Payload.String != "null" {
				payload = summarizePayload(m.Payload.String)
			}
			fmt.Printf("  %s  %-12s %s\n", formatTime(m.OccurredAt), m.EventType, payload)
		}
	}

	if len(builds) == 0 && len(mutations) == 0 {
		fmt.Println("No history found.")
	}

	return 0
}

func cmdHistoryGlobal(db *DB, project string, limit int, asJSON bool) int {
	builds, err := db.QueryRecentBuilds(project, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: history:", err)
		return 1
	}

	mutations, err := db.QueryRecentMutations(project, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko: history:", err)
		return 1
	}

	if asJSON {
		out := HistoryOutput{
			Builds:    builds,
			Mutations: mutations,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return 0
	}

	// Text output
	if len(builds) > 0 {
		fmt.Println("Recent builds:")
		for _, b := range builds {
			outcome := "in progress"
			if b.Outcome.Valid {
				outcome = b.Outcome.String
			}
			fmt.Printf("  %-12s [%-12s]  %s  %-8s %s\n",
				b.TicketID, b.Project, formatTime(b.StartedAt), outcome, b.Workflow)
		}
		fmt.Println()
	}

	if len(mutations) > 0 {
		fmt.Println("Recent events:")
		for _, m := range mutations {
			ticket := m.TicketID
			if ticket == "" {
				ticket = "-"
			}
			proj := m.Project
			if proj == "" {
				proj = "-"
			}
			payload := ""
			if m.Payload.Valid && m.Payload.String != "" && m.Payload.String != "null" {
				payload = summarizePayload(m.Payload.String)
			}
			fmt.Printf("  %s  [%-12s]  %-12s  %-12s  %s\n",
				formatTime(m.OccurredAt), proj, ticket, m.EventType, payload)
		}
	}

	if len(builds) == 0 && len(mutations) == 0 {
		fmt.Println("No history found.")
	}

	return 0
}

// formatTime formats an RFC3339 timestamp for display.
func formatTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05Z", ts)
		if err != nil {
			return ts[:min(16, len(ts))]
		}
	}
	return t.Format("2006-01-02 15:04")
}

// summarizePayload extracts key info from a JSON payload for display.
func summarizePayload(payload string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return ""
	}

	// Look for common fields
	var parts []string
	if old, ok := data["old_status"]; ok {
		if new, ok := data["new_status"]; ok {
			parts = append(parts, fmt.Sprintf("status: %v -> %v", old, new))
		}
	}
	if title, ok := data["title"]; ok {
		s := fmt.Sprintf("%v", title)
		if len(s) > 40 {
			s = s[:40] + "..."
		}
		parts = append(parts, s)
	}

	return strings.Join(parts, ", ")
}
