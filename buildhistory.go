package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BuildHistoryPath returns the JSONL build history path for a ticket.
func BuildHistoryPath(ticketsDir, id string) string {
	return filepath.Join(ticketsDir, id+".jsonl")
}

// BuildHistoryLogger appends structured JSONL events to a per-ticket
// build history file. Unlike EventLogger (which is per-build and truncated),
// this is append-only and persists across builds and ticket close.
type BuildHistoryLogger struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// OpenBuildHistory opens (or creates) the build history file for a ticket.
// The file is opened in append mode — existing events are preserved.
func OpenBuildHistory(ticketsDir, id string) (*BuildHistoryLogger, error) {
	path := BuildHistoryPath(ticketsDir, id)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &BuildHistoryLogger{file: f, path: path}, nil
}

// Path returns the file path of the build history.
func (h *BuildHistoryLogger) Path() string {
	return h.path
}

// Close closes the underlying file.
func (h *BuildHistoryLogger) Close() {
	if h.file != nil {
		h.file.Close()
	}
}

func (h *BuildHistoryLogger) emit(fields map[string]interface{}) {
	if h.file == nil {
		return
	}
	fields["ts"] = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(fields)
	if err != nil {
		return
	}
	data = append(data, '\n')
	h.mu.Lock()
	h.file.Write(data)
	h.file.Sync()
	h.mu.Unlock()
}

// BuildStart records the beginning of a build.
func (h *BuildHistoryLogger) BuildStart(ticket string) {
	h.emit(map[string]interface{}{
		"event":  "build_start",
		"ticket": ticket,
	})
}

// BuildComplete records the terminal outcome of a build.
func (h *BuildHistoryLogger) BuildComplete(ticket, outcome string) {
	h.emit(map[string]interface{}{
		"event":   "build_complete",
		"ticket":  ticket,
		"outcome": outcome,
	})
}

// NodeStart records a node beginning execution.
func (h *BuildHistoryLogger) NodeStart(ticket, workflow, node string) {
	h.emit(map[string]interface{}{
		"event":    "node_start",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
	})
}

// NodeComplete records a node finishing execution.
func (h *BuildHistoryLogger) NodeComplete(ticket, workflow, node, result string) {
	h.emit(map[string]interface{}{
		"event":    "node_complete",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
		"result":   result,
	})
}

// WorkflowStart records a workflow beginning.
func (h *BuildHistoryLogger) WorkflowStart(ticket, workflow string) {
	h.emit(map[string]interface{}{
		"event":    "workflow_start",
		"ticket":   ticket,
		"workflow": workflow,
	})
}

// NodeFail records a node execution failure with the error reason and attempt number.
func (h *BuildHistoryLogger) NodeFail(ticket, workflow, node, reason string, attempt int) {
	h.emit(map[string]interface{}{
		"event":    "node_fail",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
		"reason":   reason,
		"attempt":  attempt,
	})
}

// NodeRetry records that a failed node is being retried with the next attempt number.
func (h *BuildHistoryLogger) NodeRetry(ticket, workflow, node string, attempt int) {
	h.emit(map[string]interface{}{
		"event":    "node_retry",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
		"attempt":  attempt,
	})
}

// BuildError records a build-level error such as hook failure.
func (h *BuildHistoryLogger) BuildError(ticket, stage, reason string) {
	h.emit(map[string]interface{}{
		"event":  "build_error",
		"ticket": ticket,
		"stage":  stage,
		"reason": reason,
	})
}
