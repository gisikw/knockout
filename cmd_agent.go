package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func cmdAgent(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, `ko agent: subcommand required

Usage: ko agent <command> [arguments]

Commands:
  build <id>   Run build pipeline against a single ticket
  loop         Build all ready tickets until queue is empty
  init         Initialize pipeline config in current project
  start        Daemonize a loop (background agent)
  stop         Stop a running background agent
  status       Check if an agent is running
  report       Show summary statistics from the last agent loop run`)
		return 1
	}

	switch args[0] {
	case "build":
		return cmdAgentBuild(args[1:])
	case "loop":
		return cmdAgentLoop(args[1:])
	case "init":
		return cmdAgentInit(args[1:])
	case "start":
		return cmdAgentStart(args[1:])
	case "stop":
		return cmdAgentStop(args[1:])
	case "status":
		return cmdAgentStatus(args[1:])
	case "report":
		return cmdAgentReport(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "ko agent: unknown subcommand '%s'\n", args[0])
		return 1
	}
}

// agentPidPath returns the path to .ko/agent.pid for the given tickets dir.
func agentPidPath(ticketsDir string) string {
	return filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.pid")
}

// agentLogPath returns the path to .ko/agent.log for the given tickets dir.
func agentLogPath(ticketsDir string) string {
	return filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.log")
}

// readAgentPid reads the PID from the agent.pid file.
// Returns 0 and an error if the file doesn't exist or can't be parsed.
func readAgentPid(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid in %s: %v", path, err)
	}
	return pid, nil
}

// isProcessAlive checks if a process with the given PID is running.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// isAgentLocked checks whether the agent lock file is held by another process.
func isAgentLocked(ticketsDir string) bool {
	lockPath := filepath.Join(ProjectRoot(ticketsDir), ".ko", "agent.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false
	}
	defer f.Close()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return true // lock is held
	}
	// We got the lock — release it, nobody is running
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return false
}

func cmdAgentStart(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent start: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent start: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	pidPath := agentPidPath(ticketsDir)

	// Check for existing agent — lock is authoritative, PID file is secondary
	if isAgentLocked(ticketsDir) {
		pid, _ := readAgentPid(pidPath)
		if pid > 0 {
			fmt.Fprintf(os.Stderr, "ko agent start: agent already running (pid %d)\n", pid)
		} else {
			fmt.Fprintln(os.Stderr, "ko agent start: agent already running (lock held)")
		}
		return 1
	}

	// Check for existing agent via PID file (belt + suspenders)
	if pid, err := readAgentPid(pidPath); err == nil {
		if isProcessAlive(pid) {
			fmt.Fprintf(os.Stderr, "ko agent start: agent already running (pid %d)\n", pid)
			return 1
		}
		// Stale PID file — clean up
		os.Remove(pidPath)
	}

	// Re-exec ourselves as `ko agent loop` with any remaining flags
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent start: cannot find own executable: %v\n", err)
		return 1
	}

	logPath := agentLogPath(ticketsDir)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent start: cannot open log file: %v\n", err)
		return 1
	}

	loopArgs := append([]string{"agent", "loop"}, args...)
	cmd := exec.Command(self, loopArgs...)
	cmd.Dir = ProjectRoot(ticketsDir)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		fmt.Fprintf(os.Stderr, "ko agent start: %v\n", err)
		return 1
	}
	logFile.Close() // parent doesn't need the fd after fork

	// Write PID file
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent start: failed to write pid file: %v\n", err)
		cmd.Process.Kill()
		return 1
	}

	fmt.Printf("agent started (pid %d), logging to %s\n", cmd.Process.Pid, logPath)
	return 0
}

func cmdAgentStop(args []string) int {
	ticketsDir, _, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent stop: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent stop: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	pidPath := agentPidPath(ticketsDir)
	pid, err := readAgentPid(pidPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ko agent stop: no agent running")
		return 1
	}

	if !isProcessAlive(pid) {
		os.Remove(pidPath)
		fmt.Fprintln(os.Stderr, "ko agent stop: no agent running (stale pid file removed)")
		return 1
	}

	// Send SIGTERM first so the loop's signal handler can log it and clean up.
	// If still alive after 5 seconds, escalate to SIGKILL.
	syscall.Kill(-pid, syscall.SIGTERM)
	syscall.Kill(pid, syscall.SIGTERM)

	dead := false
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		if !isProcessAlive(pid) {
			dead = true
			break
		}
	}
	if !dead {
		syscall.Kill(-pid, syscall.SIGKILL)
		if proc, err := os.FindProcess(pid); err == nil {
			proc.Kill()
		}
	}

	os.Remove(pidPath)
	fmt.Printf("agent stopped (pid %d)\n", pid)

	// Clean up: reset any in_progress ticket and run on_fail hooks.
	// The loop's own defers should have handled this on SIGTERM, but
	// belt+suspenders in case it was escalated to SIGKILL.
	var p *Pipeline
	if configPath, err := FindPipelineConfig(ticketsDir); err == nil {
		p, _ = LoadPipeline(configPath)
	}
	cleanupAfterStop(ticketsDir, p)
	return 0
}

// cleanupAfterStop finds any in_progress ticket, resets it to open,
// and runs on_fail hooks. Called by ko agent stop after killing the process.
// Pipeline may be nil if config couldn't be loaded — hooks are skipped but
// ticket reset still happens.
func cleanupAfterStop(ticketsDir string, p *Pipeline) {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		return
	}
	for _, t := range tickets {
		if t.Status != "in_progress" {
			continue
		}
		fmt.Printf("cleanup: resetting %s to open\n", t.ID)

		if p != nil && len(p.OnFail) > 0 {
			projectRoot := ProjectRoot(ticketsDir)
			runHooks(ticketsDir, t, p.OnFail, projectRoot, projectRoot, BuildHistoryPath(ticketsDir, t.ID))
		}

		AddNote(t, "ko: reset to open (agent stopped)")
		setStatus(ticketsDir, t, "open")
	}
}

type agentStatusJSON struct {
	Provisioned bool   `json:"provisioned"`
	Running     bool   `json:"running"`
	Pid         int    `json:"pid,omitempty"`
	LastLog     string `json:"last_log,omitempty"`
}

func cmdAgentStatus(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent status: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent status: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	fs := flag.NewFlagSet("agent status", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent status: %v\n", err)
		return 1
	}

	status := agentStatusJSON{}

	// Check if pipeline config exists
	if _, err := FindPipelineConfig(ticketsDir); err != nil {
		if *jsonOutput {
			json.NewEncoder(os.Stdout).Encode(status)
		} else {
			fmt.Println("not provisioned (no .ko/config.yaml or .ko/pipeline.yml)")
		}
		return 0
	}
	status.Provisioned = true

	pidPath := agentPidPath(ticketsDir)
	pid, err := readAgentPid(pidPath)
	if err != nil {
		// No PID file — check if a lock is held (orphaned agent)
		if isAgentLocked(ticketsDir) {
			status.Running = true
			if *jsonOutput {
				json.NewEncoder(os.Stdout).Encode(status)
			} else {
				fmt.Println("running (pid unknown — lock held, pid file missing)")
			}
		} else {
			if *jsonOutput {
				json.NewEncoder(os.Stdout).Encode(status)
			} else {
				fmt.Println("not running")
			}
		}
		return 0
	}

	if isProcessAlive(pid) {
		status.Running = true
		status.Pid = pid
		logPath := agentLogPath(ticketsDir)
		if last := lastLogLine(logPath); last != "" {
			status.LastLog = last
		}

		if *jsonOutput {
			json.NewEncoder(os.Stdout).Encode(status)
		} else {
			fmt.Printf("running (pid %d)\n", pid)
			if status.LastLog != "" {
				fmt.Printf("  last: %s\n", status.LastLog)
			}
		}
	} else {
		os.Remove(pidPath)
		if *jsonOutput {
			json.NewEncoder(os.Stdout).Encode(status)
		} else {
			fmt.Println("not running (stale pid file removed)")
		}
	}
	return 0
}

// lastLogLine returns the last non-empty line from the log file.
func lastLogLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var last string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			last = line
		}
	}
	return last
}

type agentReportJSON struct {
	Timestamp        string  `json:"ts"`
	TicketsProcessed int     `json:"tickets_processed"`
	Succeeded        int     `json:"succeeded"`
	Failed           int     `json:"failed"`
	Blocked          int     `json:"blocked"`
	Decomposed       int     `json:"decomposed"`
	StopReason       string  `json:"stop_reason"`
	RuntimeSeconds   float64 `json:"runtime_seconds"`
}

func cmdAgentReport(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent report: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent report: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	fs := flag.NewFlagSet("agent report", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent report: %v\n", err)
		return 1
	}

	logPath := agentLogPath(ticketsDir)
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			if *jsonOutput {
				fmt.Println("{}")
			} else {
				fmt.Println("no runs found")
			}
			return 0
		}
		fmt.Fprintf(os.Stderr, "ko agent report: %v\n", err)
		return 1
	}
	defer f.Close()

	// Scan for the last JSONL line (lines starting with '{')
	var lastJSONLine string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "{") {
			lastJSONLine = line
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent report: error reading log: %v\n", err)
		return 1
	}

	if lastJSONLine == "" {
		if *jsonOutput {
			fmt.Println("{}")
		} else {
			fmt.Println("no runs found")
		}
		return 0
	}

	var report agentReportJSON
	if err := json.Unmarshal([]byte(lastJSONLine), &report); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent report: failed to parse summary: %v\n", err)
		return 1
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
	} else {
		fmt.Printf("Last agent loop run:\n")
		fmt.Printf("  Timestamp:  %s\n", report.Timestamp)
		fmt.Printf("  Processed:  %d tickets\n", report.TicketsProcessed)
		fmt.Printf("  Succeeded:  %d\n", report.Succeeded)
		fmt.Printf("  Failed:     %d\n", report.Failed)
		fmt.Printf("  Blocked:    %d\n", report.Blocked)
		fmt.Printf("  Decomposed: %d\n", report.Decomposed)
		fmt.Printf("  Stop reason: %s\n", report.StopReason)
		fmt.Printf("  Runtime:    %.2fs\n", report.RuntimeSeconds)
	}
	return 0
}
