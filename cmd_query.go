package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func cmdQuery(args []string) int {
	ticketsDir, _, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko query: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko query: %v\n", err)
		return 1
	}
	SortByPriorityThenModified(tickets)

	enc := json.NewEncoder(os.Stdout)
	for _, t := range tickets {
		j := ticketToJSON(t, ticketsDir)
		enc.Encode(j)
	}
	return 0
}
