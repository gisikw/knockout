package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Disposition represents a parsed decision node result.
type Disposition struct {
	Type     string   `json:"disposition"`          // continue, fail, blocked, decompose, route
	Reason   string   `json:"reason,omitempty"`     // for fail and blocked
	BlockOn  string   `json:"block_on,omitempty"`   // for blocked
	Workflow string   `json:"workflow,omitempty"`    // for route
	Subtasks []string `json:"subtasks,omitempty"`    // for decompose
}

// Valid disposition types.
var validDispositions = map[string]bool{
	"continue":  true,
	"fail":      true,
	"blocked":   true,
	"decompose": true,
	"route":     true,
}

// ExtractLastFencedJSON finds the last fenced code block in the output
// and returns its contents. Returns ("", false) if no fenced block found.
// Pure decision function.
func ExtractLastFencedJSON(output string) (string, bool) {
	lines := strings.Split(output, "\n")
	var lastBlock string
	var inBlock bool
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			// Look for opening fence: ``` or ```json
			if trimmed == "```" || trimmed == "```json" {
				inBlock = true
				current.Reset()
				continue
			}
		} else {
			// Look for closing fence
			if trimmed == "```" {
				inBlock = false
				lastBlock = current.String()
				continue
			}
			current.WriteString(line)
			current.WriteString("\n")
		}
	}

	if lastBlock == "" {
		return "", false
	}
	return strings.TrimSpace(lastBlock), true
}

// ParseDisposition parses a JSON string into a Disposition and validates it.
// Pure decision function.
func ParseDisposition(jsonStr string) (Disposition, error) {
	var d Disposition
	if err := json.Unmarshal([]byte(jsonStr), &d); err != nil {
		return Disposition{}, fmt.Errorf("invalid disposition JSON: %v", err)
	}

	if d.Type == "" {
		return Disposition{}, fmt.Errorf("disposition missing required 'disposition' field")
	}

	if !validDispositions[d.Type] {
		return Disposition{}, fmt.Errorf("unknown disposition type '%s'", d.Type)
	}

	// Validate required fields per type
	switch d.Type {
	case "route":
		if d.Workflow == "" {
			return Disposition{}, fmt.Errorf("'route' disposition missing required 'workflow' field")
		}
	case "decompose":
		if len(d.Subtasks) == 0 {
			return Disposition{}, fmt.Errorf("'decompose' disposition missing required 'subtasks' field")
		}
	}

	return d, nil
}

// DispositionSchema is the text injected into decision node prompts
// via --append-system-prompt, so the LLM knows the expected output format.
const DispositionSchema = `## Decision Node Output Format

You are a decision node. After your analysis, you MUST end your response with a fenced JSON block containing your disposition. The runner extracts the LAST fenced JSON block from your output.

Valid dispositions:

Continue to next stage:
` + "```json" + `
{"disposition": "continue"}
` + "```" + `

Fail (block ticket for human review):
` + "```json" + `
{"disposition": "fail", "reason": "Cannot implement: missing API spec"}
` + "```" + `

Blocked on another ticket:
` + "```json" + `
{"disposition": "blocked", "block_on": "ko-xxxx", "reason": "Needs auth refactor first"}
` + "```" + `

Route to a different workflow:
` + "```json" + `
{"disposition": "route", "workflow": "workflow_name"}
` + "```" + `

Decompose into subtasks:
` + "```json" + `
{"disposition": "decompose", "subtasks": ["First subtask", "Second subtask"]}
` + "```" + `

Think freely before your disposition — only the last fenced JSON block is parsed.`
