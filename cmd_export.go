package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ExportSchemaVersion identifies the shape of the ko export JSON. This is the
// contract consumed by Questbook's bulk-import endpoint. Bump it on any
// breaking change to the JSON structure below. See EXPORT_SCHEMA.md.
const ExportSchemaVersion = "1"

// KnockoutExport is the top-level export document. It is a complete dump of
// every ticket across every registered project.
type KnockoutExport struct {
	SchemaVersion string          `json:"schema_version"`
	Generator     string          `json:"generator"`
	ExportedAt    string          `json:"exported_at"`
	ProjectCount  int             `json:"project_count"`
	TicketCount   int             `json:"ticket_count"`
	Projects      []ExportProject `json:"projects"`
}

// ExportProject is one registered project and all of its tickets.
type ExportProject struct {
	Tag       string         `json:"tag"`
	Prefix    string         `json:"prefix"`
	Path      string         `json:"path"`
	IsDefault bool           `json:"is_default"`
	Hidden    bool           `json:"hidden,omitempty"`
	Tickets   []ExportTicket `json:"tickets"`
}

// ExportTicket is a single ticket with every field ko persists. deps and tags
// are always present (possibly empty) so importers have a stable contract.
type ExportTicket struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Body          string         `json:"body"`
	Status        string         `json:"status"`
	Type          string         `json:"type"`
	Priority      int            `json:"priority"`
	Assignee      string         `json:"assignee,omitempty"`
	Parent        string         `json:"parent,omitempty"`
	ExternalRef   string         `json:"external_ref,omitempty"`
	Snooze        string         `json:"snooze,omitempty"`
	Triage        string         `json:"triage,omitempty"`
	Deps          []string       `json:"deps"`
	Tags          []string       `json:"tags"`
	PlanQuestions []PlanQuestion `json:"plan_questions,omitempty"`
	Created       string         `json:"created"`
	Modified      string         `json:"modified"`
	History       []ExportEvent  `json:"history,omitempty"`
}

// ExportEvent is one entry of a ticket's mutation history (create, update, dep,
// note, ...). Payload is the raw JSON recorded with the event, if any.
type ExportEvent struct {
	OccurredAt string          `json:"occurred_at"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

func cmdExport(args []string) int {
	args = reorderArgs(args, map[string]bool{"out": true, "project": true})

	var outPath, onlyProject string
	includeHistory := true
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--out" && i+1 < len(args):
			outPath = args[i+1]
			i++
		case len(a) > 6 && a[:6] == "--out=":
			outPath = a[6:]
		case a == "--project" && i+1 < len(args):
			onlyProject = CleanTag(args[i+1])
			i++
		case len(a) > 10 && a[:10] == "--project=":
			onlyProject = CleanTag(a[10:])
		case a == "--no-history":
			includeHistory = false
		default:
			fmt.Fprintf(os.Stderr, "ko export: unexpected argument %q\n", a)
			fmt.Fprintln(os.Stderr, "usage: ko export [--out FILE] [--project TAG] [--no-history]")
			return 1
		}
	}

	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko export: cannot determine config directory")
		return 1
	}
	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko export: %v\n", err)
		return 1
	}

	var history map[string][]MutationEntry
	if includeHistory {
		if db := getShadowDB(); db != nil {
			if h, herr := db.ExportTicketHistory(); herr == nil {
				history = h
			} else {
				fmt.Fprintf(os.Stderr, "ko export: warning: could not load history: %v\n", herr)
			}
		}
	}

	// Deterministic project ordering.
	tags := make([]string, 0, len(reg.Projects))
	for tag := range reg.Projects {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	exp := KnockoutExport{
		SchemaVersion: ExportSchemaVersion,
		Generator:     "ko export " + version,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Projects:      []ExportProject{},
	}

	for _, tag := range tags {
		if onlyProject != "" && tag != onlyProject {
			continue
		}
		root := reg.Projects[tag]
		ticketsDir := resolveTicketsDir(root)
		abs, aerr := filepath.Abs(ticketsDir)
		if aerr != nil {
			abs = ticketsDir
		}

		tickets, terr := ListTickets(abs)
		if terr != nil {
			fmt.Fprintf(os.Stderr, "ko export: warning: project %q: %v\n", tag, terr)
		}
		SortByPriorityThenID(tickets)

		pj := ExportProject{
			Tag:       tag,
			Prefix:    reg.Prefixes[tag],
			Path:      root,
			IsDefault: reg.Default == tag,
			Hidden:    reg.Hidden[tag],
			Tickets:   make([]ExportTicket, 0, len(tickets)),
		}
		for _, t := range tickets {
			pj.Tickets = append(pj.Tickets, ticketToExport(t, history))
			exp.TicketCount++
		}
		exp.Projects = append(exp.Projects, pj)
	}
	exp.ProjectCount = len(exp.Projects)

	if onlyProject != "" && exp.ProjectCount == 0 {
		fmt.Fprintf(os.Stderr, "ko export: unknown project %q\n", onlyProject)
		return 1
	}

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko export: %v\n", err)
		return 1
	}
	data = append(data, '\n')

	if outPath == "" || outPath == "-" {
		os.Stdout.Write(data)
	} else {
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "ko export: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "exported %d tickets across %d projects to %s\n",
			exp.TicketCount, exp.ProjectCount, outPath)
	}
	return 0
}

// ticketToExport converts a Ticket to its export representation, attaching
// history if available.
func ticketToExport(t *Ticket, history map[string][]MutationEntry) ExportTicket {
	modified := ""
	if !t.ModTime.IsZero() {
		modified = t.ModTime.UTC().Format(time.RFC3339)
	}
	deps := t.Deps
	if deps == nil {
		deps = []string{}
	}
	tags := t.Tags
	if tags == nil {
		tags = []string{}
	}
	et := ExportTicket{
		ID:            t.ID,
		Title:         t.Title,
		Body:          t.Body,
		Status:        t.Status,
		Type:          t.Type,
		Priority:      t.Priority,
		Assignee:      t.Assignee,
		Parent:        t.Parent,
		ExternalRef:   t.ExternalRef,
		Snooze:        t.Snooze,
		Triage:        t.Triage,
		Deps:          deps,
		Tags:          tags,
		PlanQuestions: t.PlanQuestions,
		Created:       t.Created,
		Modified:      modified,
	}
	for _, m := range history[t.ID] {
		ev := ExportEvent{OccurredAt: m.OccurredAt, EventType: m.EventType}
		if m.Payload.Valid && m.Payload.String != "" {
			ev.Payload = json.RawMessage(m.Payload.String)
		}
		et.History = append(et.History, ev)
	}
	return et
}
