package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func cmdTriage(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko triage: ticket ID required")
		return 1
	}

	// Parse flags
	var blockFlag bool
	var blockReason string
	var questionsJSON string
	var answersJSON string
	var ticketID string

	for i := 0; i < len(args); i++ {
		if args[i] == "--block" {
			blockFlag = true
			// Check if next arg is a reason (not a flag, not empty)
			if i+1 < len(args) && !isFlag(args[i+1]) {
				blockReason = args[i+1]
				i++ // skip next arg
			}
		} else if args[i] == "--questions" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "ko triage: --questions flag requires a JSON argument")
				return 1
			}
			questionsJSON = args[i+1]
			i++ // skip next arg
		} else if args[i] == "--answers" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "ko triage: --answers flag requires a JSON argument")
				return 1
			}
			answersJSON = args[i+1]
			i++ // skip next arg
		} else if ticketID == "" {
			ticketID = args[i]
		}
	}

	if ticketID == "" {
		fmt.Fprintln(os.Stderr, "ko triage: ticket ID required")
		return 1
	}

	id, err := ResolveID(ticketsDir, ticketID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	// Bare invocation: show triage state
	if !blockFlag && questionsJSON == "" && answersJSON == "" {
		return showTriageState(t)
	}

	// --block: set status to blocked with optional reason
	if blockFlag {
		return handleBlock(ticketsDir, id, t, blockReason)
	}

	// --questions: add questions, implicitly block
	if questionsJSON != "" {
		return handleQuestions(ticketsDir, id, t, questionsJSON)
	}

	// --answers: answer questions, implicitly unblock when all answered
	if answersJSON != "" {
		return handleAnswers(ticketsDir, id, t, answersJSON)
	}

	return 0
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func showTriageState(t *Ticket) int {
	// Show block reason as text
	reason := ExtractBlockReason(t)
	if reason != "" {
		fmt.Printf("Block reason: %s\n\n", reason)
	} else if t.Status == "blocked" {
		fmt.Println("Status: blocked (no reason specified)")
	}

	// Show questions as JSON
	fmt.Println("Questions:")
	questions := t.PlanQuestions
	if questions == nil {
		questions = []PlanQuestion{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(questions); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: failed to encode JSON: %v\n", err)
		return 1
	}

	return 0
}

func handleBlock(ticketsDir, id string, t *Ticket, reason string) int {
	oldStatus := t.Status
	t.Status = "blocked"

	// Add note with parseable format if reason provided
	if reason != "" {
		note := fmt.Sprintf("ko: BLOCKED — %s", reason)
		AddNote(t, note)
	}

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "status", map[string]interface{}{
		"from": oldStatus,
		"to":   "blocked",
	})

	fmt.Printf("%s -> blocked\n", id)
	return 0
}

func handleQuestions(ticketsDir, id string, t *Ticket, questionsJSON string) int {
	// Parse and validate questions
	var questions []PlanQuestion
	if err := json.Unmarshal([]byte(questionsJSON), &questions); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: invalid JSON: %v\n", err)
		return 1
	}

	if err := ValidatePlanQuestions(questions); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: invalid questions: %v\n", err)
		return 1
	}

	// Set questions and status to blocked
	t.PlanQuestions = questions
	oldStatus := t.Status
	t.Status = "blocked"

	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
		return 1
	}

	EmitMutationEvent(ticketsDir, id, "status", map[string]interface{}{
		"from": oldStatus,
		"to":   "blocked",
	})

	fmt.Printf("%s -> blocked (questions added)\n", id)
	return 0
}

func handleAnswers(ticketsDir, id string, t *Ticket, answersJSON string) int {
	// Parse JSON payload
	var answers map[string]string
	if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: invalid JSON: %v\n", err)
		return 1
	}

	if len(answers) == 0 {
		fmt.Fprintln(os.Stderr, "ko triage: no answers provided in JSON")
		return 1
	}

	// Validate that ticket has plan questions
	if len(t.PlanQuestions) == 0 {
		fmt.Fprintf(os.Stderr, "ko triage: ticket %s has no plan-questions\n", id)
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
			fmt.Fprintf(os.Stderr, "ko triage: question ID %s not found in ticket %s\n", qID, id)
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

	// If all questions answered, set status to open
	if len(t.PlanQuestions) == 0 {
		t.Status = "open"
	}

	// Save ticket and emit mutation event
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko triage: %v\n", err)
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
