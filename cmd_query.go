package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ticketJSON struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Type     string   `json:"type"`
	Priority int      `json:"priority"`
	Deps     []string `json:"deps"`
	Links    []string `json:"links"`
	Created  string   `json:"created"`
	Assignee string   `json:"assignee,omitempty"`
	Parent   string   `json:"parent,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

func cmdQuery(args []string) int {
	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko query: %v\n", err)
		return 1
	}

	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko query: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(os.Stdout)
	for _, t := range tickets {
		j := ticketJSON{
			ID:       t.ID,
			Title:    t.Title,
			Status:   t.Status,
			Type:     t.Type,
			Priority: t.Priority,
			Deps:     t.Deps,
			Links:    t.Links,
			Created:  t.Created,
			Assignee: t.Assignee,
			Parent:   t.Parent,
			Tags:     t.Tags,
		}
		enc.Encode(j)
	}
	return 0
}
