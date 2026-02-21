package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// agentLockPath returns the path to .ko/agent.lock for the given tickets dir.
func agentLockPath(ticketsDir string) string {
	return filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.lock")
}

// acquireAgentLock tries to get an exclusive flock on .ko/agent.lock.
// Returns the lock file (caller must defer Close) or an error if another
// agent loop already holds it. The lock is released automatically on
// process exit, including SIGKILL.
func acquireAgentLock(ticketsDir string) (*os.File, error) {
	lockPath := agentLockPath(ticketsDir)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %v", err)
	}
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("another agent loop is already running")
	}
	return f, nil
}

func cmdAgentLoop(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}

	// Acquire exclusive lock — only one loop per project
	lockFile, err := acquireAgentLock(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}
	defer lockFile.Close()

	args = reorderArgs(args, map[string]bool{
		"max-tickets": true, "max-duration": true,
	})

	fs := flag.NewFlagSet("agent loop", flag.ContinueOnError)
	maxTickets := fs.Int("max-tickets", 0, "max tickets to process (0 = unlimited)")
	maxDuration := fs.String("max-duration", "", "max wall-clock duration (e.g. 30m, 2h)")
	quiet := fs.Bool("quiet", false, "suppress stdout; emit summary on exit")
	verbose := fs.Bool("verbose", false, "stream full agent output to stdout")
	fs.BoolVar(verbose, "v", false, "stream full agent output to stdout")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}

	configPath, err := FindPipelineConfig(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}

	p, err := LoadPipeline(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}

	config := LoopConfig{MaxTickets: *maxTickets, Quiet: *quiet, Verbose: *verbose}

	if *maxDuration != "" {
		d, err := time.ParseDuration(*maxDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko agent loop: invalid duration %q: %v\n", *maxDuration, err)
			return 1
		}
		config.MaxDuration = d
	}

	// Trap SIGTERM for graceful shutdown
	stop := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		close(stop)
	}()

	// On any exit (panic, signal, normal), reset in_progress tickets and
	// run on_fail hooks so tickets don't get stuck.
	defer cleanupAfterStop(ticketsDir, p)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "loop: panic: %v\n", r)
		}
	}()

	log := OpenEventLog()
	defer log.Close()
	result := RunLoop(ticketsDir, p, config, log, stop)

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
