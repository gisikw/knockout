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
	quiet := fs.Bool("quiet", false, "suppress stdout; emit summary on exit")

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

	config := LoopConfig{MaxTickets: *maxTickets, Quiet: *quiet}

	if *maxDuration != "" {
		d, err := time.ParseDuration(*maxDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko loop: invalid duration %q: %v\n", *maxDuration, err)
			return 1
		}
		config.MaxDuration = d
	}

	log := OpenEventLog()
	defer log.Close()
	result := RunLoop(ticketsDir, p, config, log)
	log.LoopSummary(result)

	summary := fmt.Sprintf("loop complete: %d processed (%d succeeded, %d failed, %d blocked, %d decomposed)\nstopped: %s",
		result.Processed, result.Succeeded, result.Failed, result.Blocked, result.Decomposed, result.Stopped)
	if *quiet {
		if logPath := os.Getenv("KO_EVENT_LOG"); logPath != "" {
			summary += fmt.Sprintf("\nSee %s for details", logPath)
		}
	}
	fmt.Println(summary)

	return 0
}
