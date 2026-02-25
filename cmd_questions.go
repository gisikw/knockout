package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func cmdQuestions(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko questions: %v\n", err)
		return 1
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko questions: usage: ko questions <id>")
		return 1
	}

	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko questions: %v\n", err)
		return 1
	}

	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko questions: %v\n", err)
		return 1
	}

	// Output plan-questions as JSON, or empty array if none exist
	questions := t.PlanQuestions
	if questions == nil {
		questions = []PlanQuestion{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(questions); err != nil {
		fmt.Fprintf(os.Stderr, "ko questions: failed to encode JSON: %v\n", err)
		return 1
	}

	return 0
}
