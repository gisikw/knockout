package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// MutationEvent represents a ticket mutation for the global event stream.
type MutationEvent struct {
	Timestamp string                 `json:"ts"`
	Event     string                 `json:"event"`
	Project   string                 `json:"project"`
	Ticket    string                 `json:"ticket"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// mutationEventPath returns the path to the global mutation event file.
// Uses $XDG_STATE_HOME/knockout/events.jsonl, defaulting to ~/.local/state/knockout/events.jsonl.
func mutationEventPath() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "knockout", "events.jsonl")
}

// EmitMutationEvent appends a mutation event to the global event stream.
// Best-effort: failures are silently ignored (the CLI should never fail because of event logging).
func EmitMutationEvent(ticketsDir, ticketID, event string, data map[string]interface{}) {
	path := mutationEventPath()
	if path == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}

	e := MutationEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     event,
		Project:   ProjectRoot(ticketsDir),
		Ticket:    ticketID,
		Data:      data,
	}

	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	line = append(line, '\n')

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(line)
}
