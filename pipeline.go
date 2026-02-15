package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pipeline represents a parsed .ko/pipeline.yml config (v2).
type Pipeline struct {
	Command    string                // command to invoke for prompt nodes (default: "claude")
	Model      string                // default model for prompt nodes
	MaxRetries int                   // max retries per node (default: 2)
	MaxDepth   int                   // max decomposition depth (default: 2)
	Discretion string                // low | medium | high (default: "medium")
	Workflows  map[string]*Workflow  // named workflows; "main" is the entry point
	OnSucceed  []string              // shell commands to run after all stages pass
	OnClose    []string              // shell commands to run after ticket is closed
}

// FindPipelineConfig walks up from the tickets directory looking for .ko/pipeline.yml.
func FindPipelineConfig(ticketsDir string) (string, error) {
	// .tickets is in the project root, so .ko is a sibling
	projectRoot := filepath.Dir(ticketsDir)
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
		Command:    "claude",
		MaxRetries: 2,
		MaxDepth:   2,
		Discretion: "medium",
		Workflows:  make(map[string]*Workflow),
	}

	lines := strings.Split(content, "\n")
	var section string        // "", "workflows", "on_succeed", "on_close"
	var currentWF *Workflow    // current workflow being parsed
	var currentNode *Node      // current node being parsed
	var inRoutes bool          // parsing routes list for current node

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and blank lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect top-level section headers (no indentation)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// Save any pending node/workflow
			flushNode(&currentNode, currentWF)
			flushWorkflow(&currentWF, p)
			inRoutes = false

			if trimmed == "workflows:" {
				section = "workflows"
				continue
			}
			if trimmed == "on_succeed:" {
				section = "on_succeed"
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
			case "command":
				p.Command = val
			case "model":
				p.Model = val
			case "max_retries":
				fmt.Sscanf(val, "%d", &p.MaxRetries)
			case "max_depth":
				fmt.Sscanf(val, "%d", &p.MaxDepth)
			case "discretion":
				p.Discretion = val
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
				name := strings.TrimSuffix(trimmed, ":")
				currentWF = &Workflow{Name: name}

			case indent == 4 && !strings.HasPrefix(trimmed, "-") && currentWF != nil:
				// Workflow-level property (e.g. "    model: opus")
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				if key == "model" {
					currentWF.Model = val
				}

			case strings.HasPrefix(trimmed, "- name:") && currentWF != nil:
				// New node in current workflow
				flushNode(&currentNode, currentWF)
				inRoutes = false
				_, val, _ := parseYAMLLine(strings.TrimPrefix(trimmed, "- "))
				currentNode = &Node{Name: val, MaxVisits: 1}

			case inRoutes && strings.HasPrefix(trimmed, "- ") && currentNode != nil:
				// Route entry
				route := strings.TrimPrefix(trimmed, "- ")
				route = strings.TrimSpace(route)
				currentNode.Routes = append(currentNode.Routes, route)

			case currentNode != nil:
				// Node-level property
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				inRoutes = false
				switch key {
				case "type":
					currentNode.Type = NodeType(val)
				case "prompt":
					currentNode.Prompt = val
				case "run":
					currentNode.Run = val
				case "model":
					currentNode.Model = val
				case "max_visits":
					fmt.Sscanf(val, "%d", &currentNode.MaxVisits)
				case "routes":
					inRoutes = true
					// Handle inline list: routes: [a, b, c]
					if strings.HasPrefix(val, "[") {
						currentNode.Routes = parseYAMLList(val)
						inRoutes = false
					}
				}
			}

		case "on_succeed":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnSucceed = append(p.OnSucceed, cmd)
			}
		case "on_close":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnClose = append(p.OnClose, cmd)
			}
		}
	}

	// Flush remaining
	flushNode(&currentNode, currentWF)
	flushWorkflow(&currentWF, p)

	// Validate
	if err := ValidateWorkflows(p.Workflows); err != nil {
		return nil, err
	}

	return p, nil
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
	projectRoot := filepath.Dir(ticketsDir)
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
