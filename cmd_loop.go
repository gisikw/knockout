package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func cmdLoop(args []string) int {
	args = reorderArgs(args, map[string]bool{
		"max-tickets": true, "max-duration": true,
	})

	fs := flag.NewFlagSet("loop", flag.ContinueOnError)
	maxTickets := fs.Int("max-tickets", 0, "max tickets to process (0 = unlimited)")
	maxDuration := fs.String("max-duration", "", "max wall-clock duration (e.g. 30m, 2h)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko loop: %v\n", err)
		return 1
	}

	ticketsDir, err := FindTicketsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko loop: %v\n", err)
		return 1
	}

	configPath, err := FindPipelineConfig(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko loop: %v\n", err)
		return 1
	}

	p, err := LoadPipeline(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko loop: %v\n", err)
		return 1
	}

	config := LoopConfig{MaxTickets: *maxTickets}

	if *maxDuration != "" {
		d, err := time.ParseDuration(*maxDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko loop: invalid duration %q: %v\n", *maxDuration, err)
			return 1
		}
		config.MaxDuration = d
	}

	result := RunLoop(ticketsDir, p, config)

	fmt.Printf("\nloop complete: %d processed (%d succeeded, %d failed, %d blocked, %d decomposed)\n",
		result.Processed, result.Succeeded, result.Failed, result.Blocked, result.Decomposed)
	fmt.Printf("stopped: %s\n", result.Stopped)

	return 0
}
