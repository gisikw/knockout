package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func cmdAnswer(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko answer: %v\n", err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko answer: usage: ko answer <id> '<json>'")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko answer: %v\n", err)
		return 1
	}

	// Parse JSON payload
	var answers map[string]string
	if err := json.Unmarshal([]byte(args[1]), &answers); err != nil {
		fmt.Fprintf(os.Stderr, "ko answer: invalid JSON: %v\n", err)
		return 1
	}

	if len(answers) == 0 {
		fmt.Fprintln(os.Stderr, "ko answer: no answers provided in JSON")
		return 1
	}

	// Load ticket
	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko answer: %v\n", err)
		return 1
	}

	// Validate that ticket has plan questions
	if len(t.PlanQuestions) == 0 {
		fmt.Fprintf(os.Stderr, "ko answer: ticket %s has no plan-questions\n", id)
		return 1
	}

	// Build a map of question IDs to questions for validation and note generation
	questionMap := make(map[string]PlanQuestion)
	for _, q := range t.PlanQuestions {
		questionMap[q.ID] = q
	}

	// Validate that all provided question IDs exist
	for qID := range answers {
		if _, ok := questionMap[qID]; !ok {
			fmt.Fprintf(os.Stderr, "ko answer: question ID %s not found in ticket %s\n", qID, id)
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

	// If all questions answered, clear plan-questions and set status to open
	if len(t.PlanQuestions) == 0 {
		t.Status = "open"
	}

	// Save ticket and emit mutation event
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko answer: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "answer", nil)

	// Print confirmation
	answeredCount := len(answers)
	remainingCount := len(t.PlanQuestions)
	if remainingCount == 0 {
		fmt.Printf("Answered %d question(s) for %s — all questions resolved, status set to open\n", answeredCount, id)
	} else {
		fmt.Printf("Answered %d question(s) for %s — %d question(s) remaining\n", answeredCount, id, remainingCount)
	}

	return 0
}
