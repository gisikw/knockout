package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdUpdate(args []string) int {
	// Reorder args to handle flags properly
	args = reorderArgs(args, map[string]bool{
		"d": true, "t": true, "p": true, "a": true,
		"title": true, "parent": true, "external-ref": true,
		"design": true, "acceptance": true, "tags": true,
		"questions": true, "answers": true, "status": true,
	})

	// Parse flags
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	title := fs.String("title", "", "ticket title")
	desc := fs.String("d", "", "description")
	typ := fs.String("t", "", "ticket type")
	priority := fs.Int("p", -1, "priority (0-4)")
	assignee := fs.String("a", "", "assignee")
	parent := fs.String("parent", "", "parent ticket ID")
	extRef := fs.String("external-ref", "", "external reference")
	design := fs.String("design", "", "design notes")
	acceptance := fs.String("acceptance", "", "acceptance criteria")
	tags := fs.String("tags", "", "comma-separated tags")
	questionsJSON := fs.String("questions", "", "plan questions (JSON)")
	answersJSON := fs.String("answers", "", "answers to plan questions (JSON)")
	status := fs.String("status", "", "ticket status")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}

	// Get ticket ID from positional args
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "ko update: ticket ID required")
		return 1
	}

	ticketID := fs.Arg(0)

	// Resolve project tickets directory
	ticketsDir, _, err := resolveProjectTicketsDir([]string{ticketID})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}

	// Resolve ticket ID
	id, err := ResolveID(ticketsDir, ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}

	// Load ticket
	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}

	// Track what changed for mutation event
	changed := false

	// Apply updates to ticket fields
	if *title != "" {
		t.Title = *title
		changed = true
	}
	if *desc != "" {
		t.Body += "\n" + *desc + "\n"
		changed = true
	}
	if *typ != "" {
		t.Type = *typ
		changed = true
	}
	if *priority >= 0 {
		t.Priority = *priority
		changed = true
	}
	if *assignee != "" {
		t.Assignee = *assignee
		changed = true
	}
	if *parent != "" {
		// Resolve parent ID
		parentID, err := ResolveID(ticketsDir, *parent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
			return 1
		}
		t.Parent = parentID
		changed = true
	}
	if *extRef != "" {
		t.ExternalRef = *extRef
		changed = true
	}
	if *design != "" {
		t.Body += "\n## Design\n\n" + *design + "\n"
		changed = true
	}
	if *acceptance != "" {
		t.Body += "\n## Acceptance Criteria\n\n" + *acceptance + "\n"
		changed = true
	}
	if *tags != "" {
		// Tags replace, not append
		var ticketTags []string
		for _, tag := range strings.Split(*tags, ",") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				ticketTags = append(ticketTags, trimmed)
			}
		}
		t.Tags = ticketTags
		changed = true
	}

	// Handle --questions: add questions and set status to blocked
	if *questionsJSON != "" {
		var questions []PlanQuestion
		if err := json.Unmarshal([]byte(*questionsJSON), &questions); err != nil {
			fmt.Fprintf(os.Stderr, "ko update: invalid questions JSON: %v\n", err)
			return 1
		}

		if err := ValidatePlanQuestions(questions); err != nil {
			fmt.Fprintf(os.Stderr, "ko update: invalid questions: %v\n", err)
			return 1
		}

		t.PlanQuestions = questions
		t.Status = "blocked"
		changed = true
	}

	// Handle --answers: answer questions and auto-unblock if all answered
	if *answersJSON != "" {
		var answers map[string]string
		if err := json.Unmarshal([]byte(*answersJSON), &answers); err != nil {
			fmt.Fprintf(os.Stderr, "ko update: invalid answers JSON: %v\n", err)
			return 1
		}

		if len(answers) == 0 {
			fmt.Fprintln(os.Stderr, "ko update: no answers provided in JSON")
			return 1
		}

		// Validate that ticket has plan questions
		if len(t.PlanQuestions) == 0 {
			fmt.Fprintf(os.Stderr, "ko update: ticket %s has no plan-questions\n", id)
			return 1
		}

		// Build a map of question IDs to questions for validation
		questionMap := make(map[string]PlanQuestion)
		for _, q := range t.PlanQuestions {
			questionMap[q.ID] = q
		}

		// Validate that all provided question IDs exist
		for qID := range answers {
			if _, ok := questionMap[qID]; !ok {
				fmt.Fprintf(os.Stderr, "ko update: question ID %s not found in ticket %s\n", qID, id)
				return 1
			}
		}

		// Add notes for each answered question
		for qID, answer := range answers {
			q := questionMap[qID]
			// Resolve the selected option for human-readable notes
			label := answer
			var description string
			for _, opt := range q.Options {
				if opt.Value == answer {
					label = opt.Label
					description = opt.Description
					break
				}
			}
			note := fmt.Sprintf("Question: %s\nAnswer: %s", q.Question, label)
			if description != "" {
				note += "\n" + description
			}
			AddNote(t, note)
		}

		// Remove answered questions from the slice
		var remainingQuestions []PlanQuestion
		for _, q := range t.PlanQuestions {
			if _, answered := answers[q.ID]; !answered {
				remainingQuestions = append(remainingQuestions, q)
			}
		}
		t.PlanQuestions = remainingQuestions

		// Auto-unblock: if all questions answered, set status to open
		if len(t.PlanQuestions) == 0 && t.Status == "blocked" {
			t.Status = "open"
		}

		changed = true
	}

	// Handle --status: validate and set status
	if *status != "" {
		if !ValidStatus(*status) {
			fmt.Fprintf(os.Stderr, "ko update: invalid status '%s'\nvalid statuses: %s\n", *status, strings.Join(Statuses, " "))
			return 1
		}
		t.Status = *status
		changed = true
	}

	// Save ticket if any changes were made
	if !changed {
		fmt.Fprintf(os.Stderr, "ko update: no fields specified to update\n")
		return 1
	}

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}

	// Emit mutation event
	EmitMutationEvent(ticketsDir, id, "update", map[string]interface{}{
		"ticket": id,
	})

	fmt.Printf("%s updated\n", id)
	return 0
}
