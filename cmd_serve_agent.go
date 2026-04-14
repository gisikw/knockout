package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// agentProcesses tracks running agent loop processes by project path.
var agentProcesses sync.Map // map[string]*exec.Cmd

// resolveProjectPath resolves a project identifier (path or #tag) to an absolute path.
func resolveProjectPath(project string) (string, error) {
	if project == "" {
		return "", fmt.Errorf("project required")
	}

	// If it starts with # or doesn't look like a path, treat as registry tag
	if project[0] == '#' || (project[0] != '/' && project[0] != '.') {
		tag := CleanTag(project)
		regPath := RegistryPath()
		if regPath == "" {
			return "", fmt.Errorf("cannot determine config directory")
		}
		reg, err := LoadRegistry(regPath)
		if err != nil {
			return "", err
		}
		path, ok := reg.Projects[tag]
		if !ok {
			return "", fmt.Errorf("unknown project '%s'", tag)
		}
		return path, nil
	}

	// Treat as path
	return project, nil
}

// handleAgentSpawn handles POST /agent/spawn — spawns an agent loop for a project.
func handleAgentSpawn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Project string `json:"project"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	projectPath, err := resolveProjectPath(req.Project)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Look up registry tag for the path so we can pass it to agent loop
	projectArg := req.Project
	if projectPath[0] == '/' {
		// Path was given — look up the tag in registry
		regPath := RegistryPath()
		if regPath != "" {
			if reg, err := LoadRegistry(regPath); err == nil {
				for tag, path := range reg.Projects {
					if path == projectPath {
						projectArg = tag
						break
					}
				}
			}
		}
	}

	// Check if already running
	if _, loaded := agentProcesses.Load(projectPath); loaded {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "agent already running"})
		return
	}

	// Spawn agent loop as child process
	logPath := agentLogPath(resolveTicketsDir(projectPath))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("cannot open log: %v", err)})
		return
	}

	cmd := exec.Command(os.Args[0], "agent", "loop", "--project="+projectArg)
	cmd.Dir = projectPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = cleanEnvForNesting()

	if err := cmd.Start(); err != nil {
		logFile.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to start: %v", err)})
		return
	}

	// Store process and start waiter goroutine
	agentProcesses.Store(projectPath, cmd)
	go func() {
		cmd.Wait()
		logFile.Close()
		agentProcesses.Delete(projectPath)
	}()

	// Write PID file
	ticketsDir := resolveTicketsDir(projectPath)
	pidPath := agentPidPath(ticketsDir)
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"pid": cmd.Process.Pid})
}

// handleAgentKill handles POST /agent/kill — stops a running agent.
func handleAgentKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Project string `json:"project"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	projectPath, err := resolveProjectPath(req.Project)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	val, ok := agentProcesses.Load(projectPath)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no agent running"})
		return
	}

	cmd := val.(*exec.Cmd)
	pid := cmd.Process.Pid

	// SIGTERM first, wait up to 5s, then SIGKILL
	syscall.Kill(-pid, syscall.SIGTERM)
	cmd.Process.Signal(syscall.SIGTERM)

	dead := make(chan struct{})
	go func() {
		cmd.Wait()
		close(dead)
	}()

	select {
	case <-dead:
		// Clean exit
	case <-time.After(5 * time.Second):
		syscall.Kill(-pid, syscall.SIGKILL)
		cmd.Process.Kill()
		<-dead
	}

	agentProcesses.Delete(projectPath)

	// Clean up PID file
	ticketsDir := resolveTicketsDir(projectPath)
	os.Remove(agentPidPath(ticketsDir))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"stopped": true, "pid": pid})
}

// handleAgentStatus handles GET /agent/status — checks if an agent is running.
func handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	project := r.URL.Query().Get("project")
	projectPath, err := resolveProjectPath(project)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	val, ok := agentProcesses.Load(projectPath)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"running": false})
		return
	}

	cmd := val.(*exec.Cmd)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running": true,
		"pid":     cmd.Process.Pid,
	})
}

// shutdownAgents terminates all running agent processes gracefully.
func shutdownAgents() {
	agentProcesses.Range(func(key, val interface{}) bool {
		cmd := val.(*exec.Cmd)
		pid := cmd.Process.Pid
		fmt.Fprintf(os.Stdout, "ko serve: stopping agent for %s (pid %d)\n", key, pid)

		syscall.Kill(-pid, syscall.SIGTERM)
		cmd.Process.Signal(syscall.SIGTERM)

		done := make(chan struct{})
		go func() {
			cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			cmd.Process.Kill()
		}

		agentProcesses.Delete(key)
		return true
	})
}
