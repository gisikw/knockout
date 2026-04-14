package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const serveURL = "http://localhost:19876"

// isServeRunning checks if ko serve is reachable.
func isServeRunning() bool {
	resp, err := http.Get(serveURL + "/agent/status?project=_ping")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// agentStartViaServe spawns an agent via ko serve and returns the PID.
// Returns -1 if serve isn't running or spawn fails.
func agentStartViaServe(projectRoot string) (int, error) {
	if !isServeRunning() {
		return -1, fmt.Errorf("ko serve not running (start with 'ko serve' or enable the systemd service)")
	}

	reqBody, _ := json.Marshal(map[string]string{"project": projectRoot})
	resp, err := http.Post(serveURL+"/agent/spawn", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return -1, fmt.Errorf("failed to contact serve: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusConflict {
		return -1, fmt.Errorf("agent already running")
	}
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		if json.Unmarshal(body, &errResp) == nil && errResp["error"] != "" {
			return -1, fmt.Errorf("%s", errResp["error"])
		}
		return -1, fmt.Errorf("serve returned %d", resp.StatusCode)
	}

	var result struct {
		Pid int `json:"pid"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, fmt.Errorf("invalid response from serve")
	}

	return result.Pid, nil
}

// agentStopViaServe stops an agent via ko serve.
// Returns the stopped PID, or -1 if serve couldn't stop it.
func agentStopViaServe(projectRoot string) (int, bool) {
	if !isServeRunning() {
		return -1, false
	}

	reqBody, _ := json.Marshal(map[string]string{"project": projectRoot})
	resp, err := http.Post(serveURL+"/agent/kill", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return -1, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, false
	}

	var result struct {
		Pid int `json:"pid"`
	}
	body, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &result) == nil {
		return result.Pid, true
	}
	return -1, false
}

// cmdAgentStartNew is the new HTTP-based agent start.
func cmdAgentStartNew(args []string) int {
	ticketsDir, _, err := resolveProjectTicketsDir(args)
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

	// Delegate to serve
	projectRoot := ProjectRoot(ticketsDir)
	pid, err := agentStartViaServe(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent start: %v\n", err)
		return 1
	}

	logPath := agentLogPath(ticketsDir)
	fmt.Fprintf(os.Stderr, "agent started (pid %d), logging to %s\n", pid, logPath)
	return 0
}

// cmdAgentStopNew is the new HTTP-based agent stop with fallback.
func cmdAgentStopNew(args []string) int {
	ticketsDir, _, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent stop: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent stop: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	projectRoot := ProjectRoot(ticketsDir)

	// Try to stop via serve first
	if pid, ok := agentStopViaServe(projectRoot); ok {
		fmt.Fprintf(os.Stderr, "agent stopped (pid %d)\n", pid)
		var p *Pipeline
		if configPath, err := FindPipelineConfig(ticketsDir); err == nil {
			p, _ = LoadPipeline(configPath)
		}
		cleanupAfterStop(ticketsDir, p)
		return 0
	}

	// Fallback to cmdAgentStop which does direct PID kill
	return cmdAgentStopDirect(ticketsDir)
}
