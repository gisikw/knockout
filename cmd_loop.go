package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// appendLine appends a timestamped line to a file. Best-effort; errors are silent.
func appendLine(path, line string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", line)
}

// writeHeartbeat writes PID and current timestamp to the heartbeat file.
func writeHeartbeat(path string) {
	data := strconv.Itoa(os.Getpid()) + " " + time.Now().UTC().Format(time.RFC3339) + "\n"
	os.WriteFile(path, []byte(data), 0644)
}

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

// writeAgentLogSummary appends a JSONL summary line to .ko/agent.log.
func writeAgentLogSummary(ticketsDir string, result LoopResult, elapsed time.Duration) {
	agentLogPath := filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.log")
	f, err := os.OpenFile(agentLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return // Silent failure - don't block the loop on logging errors
	}
	defer f.Close()

	summary := map[string]interface{}{
		"ts":              time.Now().UTC().Format(time.RFC3339),
		"tickets_processed": result.Processed,
		"succeeded":       result.Succeeded,
		"failed":          result.Failed,
		"blocked":         result.Blocked,
		"decomposed":      result.Decomposed,
		"stop_reason":     result.Stopped,
		"runtime_seconds": elapsed.Seconds(),
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return
	}
	data = append(data, '\n')
	f.Write(data)
}

func cmdAgentLoop(args []string) int {
	args = reorderArgs(args, map[string]bool{
		"project": true, "max-tickets": true, "max-duration": true,
	})

	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent loop: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	// Acquire exclusive lock — only one loop per project
	lockFile, err := acquireAgentLock(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent loop: %v\n", err)
		return 1
	}
	defer lockFile.Close()

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

	// Trap signals for graceful shutdown and diagnostics.
	// Broad set so we log what killed us even for unexpected signals.
	stop := make(chan struct{})
	var stopOnce sync.Once
	stopHeartbeat := func() { stopOnce.Do(func() { close(stop) }) }

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP,
		syscall.SIGQUIT, syscall.SIGPIPE,
		syscall.SIGUSR1, syscall.SIGUSR2,
	)
	agentLog := filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.log")
	go func() {
		sig := <-sigCh
		appendLine(agentLog, fmt.Sprintf("loop: received signal %s (pid %d)", sig, os.Getpid()))
		stopHeartbeat()
	}()

	// Heartbeat: write PID + timestamp to .ko/agent.heartbeat every 30s.
	// If the process dies without a clean shutdown log entry, the stale
	// heartbeat tells you when it was last alive.
	heartbeatPath := filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.heartbeat")
	heartbeatDone := make(chan struct{})
	go func() {
		defer func() { close(heartbeatDone) }()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		writeHeartbeat(heartbeatPath)
		for {
			select {
			case <-ticker.C:
				writeHeartbeat(heartbeatPath)
			case <-stop:
				os.Remove(heartbeatPath)
				return
			}
		}
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

	// Capture start time for runtime calculation
	loopStart := time.Now()
	result := RunLoop(ticketsDir, p, config, log, stop)
	elapsed := time.Since(loopStart)

	// Stop heartbeat goroutine. On signal path, stop is already closed.
	// On normal exit, we need to close it ourselves.
	stopHeartbeat()
	<-heartbeatDone

	log.LoopSummary(result)

	summary := fmt.Sprintf("loop complete: %d processed (%d succeeded, %d failed, %d blocked, %d decomposed)\nstopped: %s",
		result.Processed, result.Succeeded, result.Failed, result.Blocked, result.Decomposed, result.Stopped)
	if *quiet {
		if logPath := os.Getenv("KO_EVENT_LOG"); logPath != "" {
			summary += fmt.Sprintf("\nSee %s for details", logPath)
		}
	}
	fmt.Println(summary)

	// Append JSONL summary to .ko/agent.log
	writeAgentLogSummary(ticketsDir, result, elapsed)

	// Run on_loop_complete hooks
	if err := runLoopHooks(ticketsDir, p.OnLoopComplete, result, elapsed); err != nil {
		fmt.Fprintf(os.Stderr, "on_loop_complete hook failed: %v\n", err)
		// Don't change exit code — hook failures don't affect loop result
	}

	// Clean shutdown — remove heartbeat file
	os.Remove(heartbeatPath)

	return 0
}
