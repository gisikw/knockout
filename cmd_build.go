package main

import (
	"fmt"
	"os"
)

func cmdBuild(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ko build: ticket ID required")
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	// Resolve ticket ID
	id, err := ResolveID(ticketsDir, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	// Load ticket
	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	// Check eligibility
	depsResolved := AllDepsResolved(ticketsDir, t.Deps)
	if msg := BuildEligibility(t, depsResolved); msg != "" {
		fmt.Fprintf(os.Stderr, "ko build: %s\n", msg)
		return 1
	}

	// Load pipeline config
	configPath, err := FindPipelineConfig(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	p, err := LoadPipeline(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	// Run the build
	outcome, err := RunBuild(ticketsDir, t, p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	switch outcome {
	case OutcomeSucceed:
		fmt.Printf("SUCCEED: %s closed\n", id)
		return 0
	case OutcomeFail:
		fmt.Printf("FAIL: %s blocked\n", id)
		return 1
	case OutcomeBlocked:
		fmt.Printf("BLOCKED: %s has new dependencies\n", id)
		return 1
	case OutcomeDecompose:
		fmt.Printf("DECOMPOSE: %s split into subtasks\n", id)
		return 0
	default:
		return 0
	}
}
