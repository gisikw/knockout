package main

import (
	"encoding/json"
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
	}
}

// resolveProjectTicketsDir checks args for a #project tag and resolves it
// to that project's tickets directory via the registry. Returns the tickets
// dir and the remaining args (with the tag removed).
// If no tag is found, returns the local tickets directory.
func resolveProjectTicketsDir(args []string) (string, []string, error) {
	var tag string
	var remaining []string
	for _, a := range args {
		if strings.HasPrefix(a, "#") && len(a) > 1 && tag == "" {
			tag = CleanTag(a)
		} else {
			remaining = append(remaining, a)
		}
	}

	if tag == "" {
		ticketsDir, err := FindTicketsDir()
		return ticketsDir, remaining, err
	}

	regPath := RegistryPath()
	if regPath == "" {
		return "", remaining, fmt.Errorf("cannot determine config directory")
	}
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return "", remaining, err
	}
	projectPath, ok := reg.Projects[tag]
	if !ok {
		return "", remaining, fmt.Errorf("unknown project '#%s'", tag)
	}
	ticketsDir := resolveTicketsDir(projectPath)
	if _, err := os.Stat(ticketsDir); os.IsNotExist(err) {
		return "", remaining, fmt.Errorf("no tickets directory for '#%s' (%s)", tag, ticketsDir)
	}
	return ticketsDir, remaining, nil
}

func cmdLs(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	args = reorderArgs(args, map[string]bool{"status": true, "limit": true})

	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	statusFilter := fs.String("status", "", "filter by status")
	limit := fs.Int("limit", 0, "max tickets to show")
	jsonOutput := fs.Bool("json", false, "output as JSON array")
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
		var result []ticketJSON
		count := 0
		for _, t := range tickets {
			if *statusFilter != "" && t.Status != *statusFilter {
				continue
			}
			// Default: show non-closed tickets
			if *statusFilter == "" && t.Status == "closed" {
				continue
			}
			modified := ""
			if !t.ModTime.IsZero() {
				modified = t.ModTime.UTC().Format(time.RFC3339)
			}
			j := ticketJSON{
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
			}
			result = append(result, j)
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
			// Default: show non-closed tickets
			if *statusFilter == "" && t.Status == "closed" {
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
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}

	args = reorderArgs(args, map[string]bool{"limit": true})

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
		if IsReady(t.Status, AllDepsResolved(ticketsDir, t.Deps)) {
			ready = append(ready, t)
		}
	}

	// If local queue is non-empty, return it without cross-project checks
	if len(ready) > 0 {
		SortByPriorityThenModified(ready)
		if *jsonOutput {
			var result []ticketJSON
			count := 0
			for _, t := range ready {
				modified := ""
				if !t.ModTime.IsZero() {
					modified = t.ModTime.UTC().Format(time.RFC3339)
				}
				j := ticketJSON{
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
				}
				result = append(result, j)
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
		if IsReady(t.Status, AllDepsResolvedWith(t.Deps, lookup)) {
			if *jsonOutput {
				modified := ""
				if !t.ModTime.IsZero() {
					modified = t.ModTime.UTC().Format(time.RFC3339)
				}
				j := ticketJSON{
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
					HasUnresolvedDep: !AllDepsResolvedWith(t.Deps, lookup),
					PlanQuestions:    t.PlanQuestions,
				}
				result := []ticketJSON{j}
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
