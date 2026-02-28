package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type ticketJSON struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Status            string         `json:"status"`
	Type              string         `json:"type"`
	Priority          int            `json:"priority"`
	Deps              []string       `json:"deps"`
	Created           string         `json:"created"`
	Modified          string         `json:"modified"`
	Assignee          string         `json:"assignee,omitempty"`
	Parent            string         `json:"parent,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Description       string         `json:"description,omitempty"`
	HasUnresolvedDep  bool           `json:"hasUnresolvedDep"`
	PlanQuestions     []PlanQuestion `json:"plan-questions,omitempty"`
	Snooze            string         `json:"snooze,omitempty"`
	Triage            string         `json:"triage,omitempty"`
}

// ticketToJSON converts a Ticket to ticketJSON format.
func ticketToJSON(t *Ticket, ticketsDir string) ticketJSON {
	modified := ""
	if !t.ModTime.IsZero() {
		modified = t.ModTime.UTC().Format(time.RFC3339)
	}
	return ticketJSON{
		ID:               t.ID,
		Title:            t.Title,
		Status:           t.Status,
		Type:             t.Type,
		Priority:         t.Priority,
		Deps:             t.Deps,
		Created:          t.Created,
		Modified:         modified,
		Assignee:         t.Assignee,
		Parent:           t.Parent,
		Tags:             t.Tags,
		Description:      t.Body,
		HasUnresolvedDep: !AllDepsResolved(ticketsDir, t.Deps),
		PlanQuestions:    t.PlanQuestions,
		Snooze:           t.Snooze,
		Triage:           t.Triage,
	}
}

// resolveProjectTicketsDir checks args for a --project flag or #tag shorthand
// and resolves it to that project's tickets directory via the registry.
// Returns the tickets dir and the remaining args (with the flag removed).
// If no --project flag or #tag is found, returns the local tickets directory.
// If both #tag and --project are provided, --project takes precedence.
func resolveProjectTicketsDir(args []string) (string, []string, error) {
	// Manually parse --project and #tag to avoid consuming other flags
	var projectTag string
	var remaining []string
	var foundHashTag bool

	// First pass: scan for positional #tag args (before we process --project)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "#") && !foundHashTag {
			// This is a #tag shorthand
			projectTag = CleanTag(arg)
			foundHashTag = true
			// Don't add to remaining (consume it)
		} else {
			remaining = append(remaining, arg)
		}
	}

	// Second pass: check for explicit --project flag (overrides #tag)
	var finalRemaining []string
	for i := 0; i < len(remaining); i++ {
		arg := remaining[i]
		if strings.HasPrefix(arg, "--project=") {
			projectTag = strings.TrimPrefix(arg, "--project=")
			// Don't add to finalRemaining (consume it)
		} else if arg == "--project" && i+1 < len(remaining) {
			projectTag = remaining[i+1]
			i++ // skip next arg
			// Don't add either to finalRemaining (consume both)
		} else {
			finalRemaining = append(finalRemaining, arg)
		}
	}
	remaining = finalRemaining

	if projectTag == "" {
		ticketsDir, err := FindTicketsDir()
		if err != nil && !errors.Is(err, ErrNoLocalProject) {
			return "", remaining, err
		}
		return ticketsDir, remaining, nil
	}

	regPath := RegistryPath()
	if regPath == "" {
		return "", remaining, fmt.Errorf("cannot determine config directory")
	}
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return "", remaining, err
	}
	projectPath, ok := reg.Projects[projectTag]
	if !ok {
		return "", remaining, fmt.Errorf("unknown project '%s'", projectTag)
	}
	ticketsDir := resolveTicketsDir(projectPath)
	if _, err := os.Stat(ticketsDir); os.IsNotExist(err) {
		return "", remaining, fmt.Errorf("no tickets directory for project '%s' (%s)", projectTag, ticketsDir)
	}
	return ticketsDir, remaining, nil
}

func cmdLs(args []string) int {
	args = reorderArgs(args, map[string]bool{"project": true, "status": true, "limit": true})

	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko ls: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	statusFilter := fs.String("status", "", "filter by status")
	limit := fs.Int("limit", 0, "max tickets to show")
	jsonOutput := fs.Bool("json", false, "output as JSON array")
	allTickets := fs.Bool("all", false, "include closed tickets")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}
	SortByPriorityThenModified(tickets)

	if *jsonOutput {
		result := make([]ticketJSON, 0)
		count := 0
		for _, t := range tickets {
			if *statusFilter != "" && t.Status != *statusFilter {
				continue
			}
			// Default: show non-closed tickets (unless --all is set)
			if *statusFilter == "" && !*allTickets && t.Status == "closed" {
				continue
			}
			result = append(result, ticketToJSON(t, ticketsDir))
			count++
			if *limit > 0 && count >= *limit {
				break
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	} else {
		count := 0
		for _, t := range tickets {
			if *statusFilter != "" && t.Status != *statusFilter {
				continue
			}
			// Default: show non-closed tickets (unless --all is set)
			if *statusFilter == "" && !*allTickets && t.Status == "closed" {
				continue
			}
			line := fmt.Sprintf("%s [%s] (p%d) %s", t.ID, t.Status, t.Priority, t.Title)
			if len(t.Deps) > 0 {
				line += fmt.Sprintf(" <- [%s]", strings.Join(t.Deps, ", "))
			}
			fmt.Println(line)
			count++
			if *limit > 0 && count >= *limit {
				break
			}
		}
	}
	return 0
}

func cmdReady(args []string) int {
	args = reorderArgs(args, map[string]bool{"project": true, "limit": true})

	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko ready: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "max tickets to show")
	jsonOutput := fs.Bool("json", false, "output as JSON array")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}

	// Local ready queue (deps resolved locally only)
	var ready []*Ticket
	for _, t := range tickets {
		if IsReady(t.Status, AllDepsResolved(ticketsDir, t.Deps)) && !IsSnoozed(t.Snooze, time.Now()) && t.Triage == "" {
			ready = append(ready, t)
		}
	}

	// If local queue is non-empty, return it without cross-project checks
	if len(ready) > 0 {
		SortByPriorityThenModified(ready)
		if *jsonOutput {
			result := make([]ticketJSON, 0)
			count := 0
			for _, t := range ready {
				result = append(result, ticketToJSON(t, ticketsDir))
				count++
				if *limit > 0 && count >= *limit {
					break
				}
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(result)
		} else {
			count := 0
			for _, t := range ready {
				fmt.Printf("%s [%s] (p%d) %s\n", t.ID, t.Status, t.Priority, t.Title)
				count++
				if *limit > 0 && count >= *limit {
					break
				}
			}
		}
		return 0
	}

	// Local queue empty — check cross-project deps (short-circuit on first)
	regPath := RegistryPath()
	if regPath == "" {
		return 0
	}
	reg, err := LoadRegistry(regPath)
	if err != nil || len(reg.Projects) == 0 {
		return 0
	}

	lookup := CrossProjectLookup(ticketsDir, reg)
	for _, t := range tickets {
		if IsReady(t.Status, AllDepsResolvedWith(t.Deps, lookup)) && !IsSnoozed(t.Snooze, time.Now()) && t.Triage == "" {
			if *jsonOutput {
				result := []ticketJSON{ticketToJSON(t, ticketsDir)}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(result)
			} else {
				fmt.Printf("%s [%s] (p%d) %s\n", t.ID, t.Status, t.Priority, t.Title)
			}
			return 0 // short-circuit: one ticket at a time
		}
	}

	return 0
}

func cmdTriage(args []string) int {
	args = reorderArgs(args, map[string]bool{"project": true, "limit": true})

	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko triage: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	fs := flag.NewFlagSet("triage", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "max tickets to show")
	jsonOutput := fs.Bool("json", false, "output as JSON array")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	if fs.NArg() >= 2 {
		return cmdUpdate([]string{fs.Arg(0), "--triage=" + strings.Join(fs.Args()[1:], " ")})
	}
	if fs.NArg() == 1 {
		fmt.Fprintln(os.Stderr, "ko triage: instructions required")
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	var triaged []*Ticket
	for _, t := range tickets {
		if t.Triage != "" {
			triaged = append(triaged, t)
		}
	}

	SortByPriorityThenModified(triaged)

	if *jsonOutput {
		result := make([]ticketJSON, 0)
		count := 0
		for _, t := range triaged {
			result = append(result, ticketToJSON(t, ticketsDir))
			count++
			if *limit > 0 && count >= *limit {
				break
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	} else {
		count := 0
		for _, t := range triaged {
			fmt.Printf("%s [%s] (p%d) %s \u2014 triage: %s\n", t.ID, t.Status, t.Priority, t.Title, t.Triage)
			count++
			if *limit > 0 && count >= *limit {
				break
			}
		}
	}
	return 0
}
