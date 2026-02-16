package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// EventLogger writes structured JSONL events to a file specified by
// the KO_EVENT_LOG environment variable. If the variable is not set,
// all methods are no-ops.
type EventLogger struct {
	mu   sync.Mutex
	file *os.File
}

// OpenEventLog creates a new EventLogger. If KO_EVENT_LOG is not set,
// returns a no-op logger. The log file is truncated on open.
func OpenEventLog() *EventLogger {
	path := os.Getenv("KO_EVENT_LOG")
	if path == "" {
		return &EventLogger{}
	}

	f, err := os.Create(path) // truncate + create
	if err != nil {
		return &EventLogger{}
	}
	return &EventLogger{file: f}
}

// Close closes the underlying file, if any.
func (l *EventLogger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *EventLogger) emit(fields map[string]interface{}) {
	if l.file == nil {
		return
	}
	fields["ts"] = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(fields)
	if err != nil {
		return
	}
	data = append(data, '\n')
	l.mu.Lock()
	l.file.Write(data)
	l.file.Sync()
	l.mu.Unlock()
}

// WorkflowStart logs the start of a workflow for a ticket.
func (l *EventLogger) WorkflowStart(ticket, workflow string) {
	l.emit(map[string]interface{}{
		"event":    "workflow_start",
		"ticket":   ticket,
		"workflow": workflow,
	})
}

// NodeStart logs the start of a node execution.
func (l *EventLogger) NodeStart(ticket, workflow, node string) {
	l.emit(map[string]interface{}{
		"event":    "node_start",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
	})
}

// NodeComplete logs the completion of a node.
func (l *EventLogger) NodeComplete(ticket, workflow, node, result string) {
	l.emit(map[string]interface{}{
		"event":    "node_complete",
		"ticket":   ticket,
		"workflow": workflow,
		"node":     node,
		"result":   result,
	})
}

// WorkflowComplete logs the terminal outcome of a workflow.
func (l *EventLogger) WorkflowComplete(ticket, outcome string) {
	l.emit(map[string]interface{}{
		"event":   "workflow_complete",
		"ticket":  ticket,
		"outcome": outcome,
	})
}

// LoopTicketStart logs that the loop is starting a build for a ticket.
func (l *EventLogger) LoopTicketStart(ticket, title string) {
	l.emit(map[string]interface{}{
		"event":  "loop_ticket_start",
		"ticket": ticket,
		"title":  title,
	})
}

// LoopTicketComplete logs the outcome of a single ticket in the loop.
func (l *EventLogger) LoopTicketComplete(ticket, outcome string) {
	l.emit(map[string]interface{}{
		"event":   "loop_ticket_complete",
		"ticket":  ticket,
		"outcome": outcome,
	})
}

// LoopSummary logs the final summary of a loop run.
func (l *EventLogger) LoopSummary(result LoopResult) {
	l.emit(map[string]interface{}{
		"event":       "loop_summary",
		"processed":   result.Processed,
		"succeeded":   result.Succeeded,
		"failed":      result.Failed,
		"blocked":     result.Blocked,
		"decomposed":  result.Decomposed,
		"stop_reason": result.Stopped,
	})
}
