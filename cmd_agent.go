package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
  status       Check if an agent is running`)
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

	// Kill immediately — don't wait for the current build to finish
	killProcessGroup(pid)
	os.Remove(pidPath)
	fmt.Printf("agent stopped (pid %d)\n", pid)

	// Clean up: reset any in_progress ticket and run on_fail hooks
	var p *Pipeline
	if configPath, err := FindPipelineConfig(ticketsDir); err == nil {
		p, _ = LoadPipeline(configPath)
	}
	cleanupAfterStop(ticketsDir, p)
	return 0
}

// killProcessGroup sends SIGKILL to the process group led by pid, killing
// the agent loop and any child processes (claude, sh hooks, etc).
func killProcessGroup(pid int) {
	// Try process group kill first (negative pid = kill the group)
	syscall.Kill(-pid, syscall.SIGKILL)
	// Also kill the process directly in case setsid didn't work
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Kill()
	}
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
			runHooks(ticketsDir, t, p.OnFail, projectRoot, projectRoot)
		}

		AddNote(t, "ko: reset to open (agent stopped)")
		setStatus(ticketsDir, t, "open")
	}
}

func cmdAgentStatus(args []string) int {
	ticketsDir, _, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent status: %v\n", err)
		return 1
	}

	// Check if pipeline config exists
	if _, err := FindPipelineConfig(ticketsDir); err != nil {
		fmt.Println("not provisioned (no .ko/pipeline.yml)")
		return 0
	}

	pidPath := agentPidPath(ticketsDir)
	pid, err := readAgentPid(pidPath)
	if err != nil {
		// No PID file — check if a lock is held (orphaned agent)
		if isAgentLocked(ticketsDir) {
			fmt.Println("running (pid unknown — lock held, pid file missing)")
		} else {
			fmt.Println("not running")
		}
		return 0
	}

	if isProcessAlive(pid) {
		fmt.Printf("running (pid %d)\n", pid)
		logPath := agentLogPath(ticketsDir)
		if last := lastLogLine(logPath); last != "" {
			fmt.Printf("  last: %s\n", last)
		}
	} else {
		os.Remove(pidPath)
		fmt.Println("not running (stale pid file removed)")
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
