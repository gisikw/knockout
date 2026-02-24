package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Pipeline represents a parsed .ko/pipeline.yml config (v2).
type Pipeline struct {
	Agent        string                // agent adapter name: claude | cursor (default: "claude")
	Command      string                // raw command override (mutually exclusive with agent)
	AllowAll     bool                  // maps to --dangerously-skip-permissions, --force, etc.
	AllowedTools []string              // list of tool names to auto-allow (e.g., Read, Write, Bash)
	Model        string                // default model for prompt nodes
	MaxRetries   int                   // max retries per node (default: 2)
	MaxDepth     int                   // max decomposition depth (default: 2)
	Discretion   string                // low | medium | high (default: "medium")
	StepTimeout  string                // default timeout for all nodes (e.g., "15m", "1h30m")
	Workflows    map[string]*Workflow  // named workflows; "main" is the entry point
	OnSucceed    []string              // shell commands to run after all stages pass
	OnFail       []string              // shell commands to run on build failure
	OnClose      []string              // shell commands to run after ticket is closed

	agentExplicit bool              // true if agent: was explicitly set in config
}

// Adapter returns the AgentAdapter for this pipeline config.
// If command: is set, returns a RawCommandAdapter.
// Otherwise, looks up the agent by name.
func (p *Pipeline) Adapter() AgentAdapter {
	if p.Command != "" {
		return &RawCommandAdapter{Command: p.Command}
	}
	adapter := LookupAdapter(p.Agent)
	if adapter == nil {
		// Shouldn't happen after validation, but fallback to raw
		return &RawCommandAdapter{Command: p.Agent}
	}
	return adapter
}

// FindPipelineConfig walks up from the tickets directory looking for .ko/pipeline.yml.
func FindPipelineConfig(ticketsDir string) (string, error) {
	projectRoot := ProjectRoot(ticketsDir)
	candidate := filepath.Join(projectRoot, ".ko", "pipeline.yml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("no pipeline config found (expected .ko/pipeline.yml)")
}

// LoadPipeline reads and parses a pipeline config file.
func LoadPipeline(path string) (*Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParsePipeline(string(data))
}

// ParsePipeline parses pipeline YAML content (v2 format).
// Uses the same minimal YAML approach as ticket parsing — no external deps.
func ParsePipeline(content string) (*Pipeline, error) {
	p := &Pipeline{
		Agent:      "claude",
		MaxRetries: 2,
		MaxDepth:   2,
		Discretion: "medium",
		Workflows:  make(map[string]*Workflow),
	}

	lines := strings.Split(content, "\n")
	var section string        // "", "workflows", "on_succeed", "on_fail", "on_close"
	var currentWF *Workflow    // current workflow being parsed
	var currentNode *Node      // current node being parsed
	var inRoutes bool          // parsing routes list for current node
	var inSkills bool          // parsing skills list for current node
	var inAllowedTools bool    // parsing allowed_tools list
	var inPrompt bool          // parsing inline prompt content
	var promptIndent int       // indentation level of prompt: line
	var promptLines []string   // accumulated prompt lines

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and blank lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect top-level section headers (no indentation)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// Flush any pending inline prompt
			if inPrompt && currentNode != nil {
				currentNode.Prompt = strings.Join(promptLines, "\n")
				inPrompt = false
				promptLines = nil
			}
			// Save any pending node/workflow
			flushNode(&currentNode, currentWF)
			flushWorkflow(&currentWF, p)
			inRoutes = false
			inSkills = false
			inAllowedTools = false

			if trimmed == "workflows:" {
				section = "workflows"
				continue
			}
			if trimmed == "on_succeed:" {
				section = "on_succeed"
				continue
			}
			if trimmed == "on_fail:" {
				section = "on_fail"
				continue
			}
			if trimmed == "on_close:" {
				section = "on_close"
				continue
			}

			// Top-level scalars
			key, val, ok := parseYAMLLine(trimmed)
			if !ok {
				continue
			}
			switch key {
			case "agent":
				p.Agent = val
				p.agentExplicit = true
			case "command":
				p.Command = val
			case "allow_all_tool_calls":
				p.AllowAll = val == "true"
			case "model":
				p.Model = val
			case "max_retries":
				fmt.Sscanf(val, "%d", &p.MaxRetries)
			case "max_depth":
				fmt.Sscanf(val, "%d", &p.MaxDepth)
			case "discretion":
				p.Discretion = val
			case "step_timeout":
				p.StepTimeout = val
			case "allowed_tools":
				// Handle inline list: allowed_tools: [a, b, c]
				if strings.HasPrefix(val, "[") {
					p.AllowedTools = parseYAMLList(val)
				} else {
					// Multiline list will be handled in section parsing
					section = "allowed_tools"
					continue
				}
			}
			section = ""
			continue
		}

		// Section content
		switch section {
		case "workflows":
			indent := countIndent(line)
			switch {
			case indent == 2 && strings.HasSuffix(trimmed, ":"):
				// Workflow name header (e.g. "  main:")
				flushNode(&currentNode, currentWF)
				flushWorkflow(&currentWF, p)
				inRoutes = false
				inSkills = false
				inAllowedTools = false
				name := strings.TrimSuffix(trimmed, ":")
				currentWF = &Workflow{Name: name}

			case indent == 4 && !strings.HasPrefix(trimmed, "-") && currentWF != nil:
				// Workflow-level property (e.g. "    model: opus")
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				switch key {
				case "model":
					currentWF.Model = val
				case "allow_all_tool_calls":
					v := val == "true"
					currentWF.AllowAll = &v
				case "allowed_tools":
					inAllowedTools = true
					// Handle inline list: allowed_tools: [a, b, c]
					if strings.HasPrefix(val, "[") {
						currentWF.AllowedTools = parseYAMLList(val)
						inAllowedTools = false
					}
				}

			case strings.HasPrefix(trimmed, "- name:") && currentWF != nil:
				// New node in current workflow
				// Flush any pending inline prompt first
				if inPrompt && currentNode != nil {
					currentNode.Prompt = strings.Join(promptLines, "\n")
					inPrompt = false
					promptLines = nil
				}
				flushNode(&currentNode, currentWF)
				inRoutes = false
				inSkills = false
				inAllowedTools = false
				_, val, _ := parseYAMLLine(strings.TrimPrefix(trimmed, "- "))
				currentNode = &Node{Name: val, MaxVisits: 1}

			case inAllowedTools && strings.HasPrefix(trimmed, "- ") && currentWF != nil && currentNode == nil:
				// Workflow-level allowed_tools entry
				tool := strings.TrimPrefix(trimmed, "- ")
				tool = strings.TrimSpace(tool)
				currentWF.AllowedTools = append(currentWF.AllowedTools, tool)

			case inRoutes && strings.HasPrefix(trimmed, "- ") && currentNode != nil:
				// Route entry
				route := strings.TrimPrefix(trimmed, "- ")
				route = strings.TrimSpace(route)
				currentNode.Routes = append(currentNode.Routes, route)

			case inSkills && strings.HasPrefix(trimmed, "- ") && currentNode != nil:
				// Skill entry
				skill := strings.TrimPrefix(trimmed, "- ")
				skill = strings.TrimSpace(skill)
				currentNode.Skills = append(currentNode.Skills, skill)

			case inAllowedTools && strings.HasPrefix(trimmed, "- ") && currentNode != nil:
				// Node-level allowed_tools entry
				tool := strings.TrimPrefix(trimmed, "- ")
				tool = strings.TrimSpace(tool)
				currentNode.AllowedTools = append(currentNode.AllowedTools, tool)

			case inPrompt && currentNode != nil:
				// Accumulating inline prompt lines
				indent := countIndent(line)
				if indent <= promptIndent {
					// End of prompt block — line is at same or lesser indent
					currentNode.Prompt = strings.Join(promptLines, "\n")
					inPrompt = false
					promptLines = nil
					// Re-process this line as a node property
					key, val, ok := parseYAMLLine(trimmed)
					if ok {
						applyNodeProperty(currentNode, key, val, &inRoutes, &inSkills, &inAllowedTools)
					}
				} else {
					// Strip common indentation and accumulate
					stripped := line
					if len(line) > promptIndent+2 {
						stripped = line[promptIndent+2:]
					}
					promptLines = append(promptLines, stripped)
				}

			case currentNode != nil:
				// Node-level property
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				inRoutes = false
				inSkills = false
				inAllowedTools = false
				if key == "prompt" && val == "|" {
					// Start of inline prompt
					inPrompt = true
					promptIndent = countIndent(line)
					promptLines = nil
				} else {
					applyNodeProperty(currentNode, key, val, &inRoutes, &inSkills, &inAllowedTools)
				}
			}

		case "on_succeed":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnSucceed = append(p.OnSucceed, cmd)
			}
		case "on_fail":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnFail = append(p.OnFail, cmd)
			}
		case "on_close":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnClose = append(p.OnClose, cmd)
			}
		case "allowed_tools":
			if strings.HasPrefix(trimmed, "- ") {
				tool := strings.TrimPrefix(trimmed, "- ")
				tool = strings.TrimSpace(tool)
				p.AllowedTools = append(p.AllowedTools, tool)
			}
		}
	}

	// Flush any pending inline prompt
	if inPrompt && currentNode != nil {
		currentNode.Prompt = strings.Join(promptLines, "\n")
	}
	// Flush remaining
	flushNode(&currentNode, currentWF)
	flushWorkflow(&currentWF, p)

	// Validate: agent and command are mutually exclusive.
	// If command: is set without an explicit agent:, clear the default agent.
	if p.Command != "" && p.Agent == "claude" && !p.agentExplicit {
		p.Agent = ""
	}
	if p.Agent != "" && p.Command != "" {
		return nil, fmt.Errorf("pipeline config cannot set both 'agent' and 'command'")
	}
	if err := ValidateWorkflows(p.Workflows); err != nil {
		return nil, err
	}

	return p, nil
}

// applyNodeProperty applies a parsed key-value pair to a node.
func applyNodeProperty(node *Node, key, val string, inRoutes *bool, inSkills *bool, inAllowedTools *bool) {
	*inRoutes = false
	*inSkills = false
	*inAllowedTools = false
	switch key {
	case "type":
		node.Type = NodeType(val)
	case "prompt":
		node.Prompt = val
	case "run":
		node.Run = val
	case "model":
		node.Model = val
	case "allow_all_tool_calls":
		v := val == "true"
		node.AllowAll = &v
	case "max_visits":
		fmt.Sscanf(val, "%d", &node.MaxVisits)
	case "timeout":
		node.Timeout = val
	case "routes":
		*inRoutes = true
		// Handle inline list: routes: [a, b, c]
		if strings.HasPrefix(val, "[") {
			node.Routes = parseYAMLList(val)
			*inRoutes = false
		}
	case "skills":
		*inSkills = true
		// Handle inline list: skills: [a, b, c]
		if strings.HasPrefix(val, "[") {
			node.Skills = parseYAMLList(val)
			*inSkills = false
		}
	case "allowed_tools":
		*inAllowedTools = true
		// Handle inline list: allowed_tools: [a, b, c]
		if strings.HasPrefix(val, "[") {
			node.AllowedTools = parseYAMLList(val)
			*inAllowedTools = false
		}
	case "skill":
		node.Skill = val
	}
}

// flushNode saves the current node to the current workflow and clears it.
func flushNode(node **Node, wf *Workflow) {
	if *node == nil || wf == nil {
		return
	}
	wf.Nodes = append(wf.Nodes, **node)
	*node = nil
}

// flushWorkflow saves the current workflow to the pipeline and clears it.
func flushWorkflow(wf **Workflow, p *Pipeline) {
	if *wf == nil {
		return
	}
	p.Workflows[(*wf).Name] = *wf
	*wf = nil
}

// countIndent returns the number of leading spaces in a line.
func countIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 2 // treat tab as 2 spaces
		} else {
			break
		}
	}
	return count
}

// LoadPromptFile reads a prompt file from .ko/prompts/<name>.
func LoadPromptFile(ticketsDir, name string) (string, error) {
	projectRoot := ProjectRoot(ticketsDir)
	path := filepath.Join(projectRoot, ".ko", "prompts", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt file '%s' not found", name)
	}
	return string(data), nil
}

// DiscretionGuidance returns instructional text for a discretion level.
func DiscretionGuidance(level string) string {
	switch level {
	case "high":
		return "You have HIGH discretion. Make reasonable assumptions and keep moving. " +
			"Note your assumptions but don't block on ambiguity that can be resolved " +
			"by looking at the code. Bias toward shipping."
	case "low":
		return "You have LOW discretion. Be conservative. Flag anything ambiguous. " +
			"If there are multiple reasonable approaches, signal FAIL rather " +
			"than guessing. Prefer asking over assuming."
	default: // medium
		return "You have MEDIUM discretion. Investigate the codebase before flagging issues. " +
			"Make assumptions for obvious choices, but flag genuine architectural decisions."
	}
}

// parseTimeout parses a duration string and returns the corresponding time.Duration.
// Returns 15 minutes if the input is empty.
// Returns an error if the format is invalid.
func parseTimeout(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 15 * time.Minute, nil
	}
	return time.ParseDuration(durationStr)
}
