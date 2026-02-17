package main

import (
	"flag"
	"fmt"
	"os"
)

func cmdBuild(args []string) int {
	args = reorderArgs(args, map[string]bool{})

	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	quiet := fs.Bool("quiet", false, "suppress stdout; emit summary on exit")
	verbose := fs.Bool("verbose", false, "stream full agent output to stdout")
	fs.BoolVar(verbose, "v", false, "stream full agent output to stdout")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "ko build: ticket ID required")
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	// Resolve ticket ID
	id, err := ResolveID(ticketsDir, fs.Arg(0))
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
	log := OpenEventLog()
	defer log.Close()
	outcome, err := RunBuild(ticketsDir, t, p, log, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko build: %v\n", err)
		return 1
	}

	if *quiet {
		summary := fmt.Sprintf("build: %s %s", id, outcomeString(outcome))
		if logPath := os.Getenv("KO_EVENT_LOG"); logPath != "" {
			summary += fmt.Sprintf(". See %s for details", logPath)
		}
		fmt.Println(summary)
	} else {
		switch outcome {
		case OutcomeSucceed:
			fmt.Printf("SUCCEED: %s closed\n", id)
		case OutcomeFail:
			fmt.Printf("FAIL: %s blocked\n", id)
		case OutcomeBlocked:
			fmt.Printf("BLOCKED: %s has new dependencies\n", id)
		case OutcomeDecompose:
			fmt.Printf("DECOMPOSE: %s split into subtasks\n", id)
		}
	}

	switch outcome {
	case OutcomeFail, OutcomeBlocked:
		return 1
	default:
		return 0
	}
}
