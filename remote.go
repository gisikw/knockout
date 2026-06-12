package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// localOnlyCommands are commands that should never proxy to a remote server.
var localOnlyCommands = map[string]bool{
	"serve":   true, // is the server
	"agent":   true, // pipeline execution is local
	"help":    true,
	"--help":  true,
	"-h":      true,
	"version": true,
	"--version": true,
	"-v":      true,
	"import":  true, // not in serve whitelist
}

// isRemoteCommand returns true if the command should proxy to a remote server.
func isRemoteCommand(cmd string) bool {
	return !localOnlyCommands[cmd]
}

// remoteExec sends a command to a remote ko serve instance and prints the result.
// Returns an exit code: 0 on success, 1 on failure.
func remoteExec(server string, argv []string) int {
	// Build request body
	body, err := json.Marshal(struct {
		Argv []string `json:"argv"`
	}{Argv: argv})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: failed to marshal request: %v\n", err)
		return 1
	}

	// POST to server
	url := strings.TrimRight(server, "/") + "/ko"
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: server unreachable: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: failed to read response: %v\n", err)
		return 1
	}

	if resp.StatusCode == http.StatusOK {
		os.Stdout.Write(respBody)
		return 0
	}

	// Try to parse JSON error response from ko serve
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
		fmt.Fprint(os.Stderr, errResp.Error)
		return 1
	}

	// Fallback: print raw body
	fmt.Fprintf(os.Stderr, "ko: server returned %d: %s\n", resp.StatusCode, string(respBody))
	return 1
}
