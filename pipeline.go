package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config represents a unified .ko/config.yaml file containing both project
// settings and pipeline configuration.
type Config struct {
	Project    ProjectConfig // project-level settings (prefix, etc.)
	Pipeline   Pipeline      // pipeline configuration
	Summarizer string        // command to summarize long titles (overrides global)
}

// ProjectConfig holds project-level settings from the config.yaml project: section.
type ProjectConfig struct {
	Prefix string // ticket ID prefix (e.g., "ko")
}

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
	StepTimeout      string                // default timeout for all nodes (e.g., "15m", "1h30m")
	// RequireCleanTree requires working tree to be clean (no uncommitted changes outside .ko/) before build starts
	RequireCleanTree bool
	// AutoTriage automatically runs ko agent triage when a ticket is created or updated with a triage field set
	AutoTriage bool
	// AutoAgent automatically starts the agent loop when a ticket is created or becomes actionable
	AutoAgent bool
	// Workers is the number of parallel ticket builds (default: 1 = sequential).
	// When > 1, each ticket runs in its own git worktree for filesystem isolation.
	Workers int
	Workflows        map[string]*Workflow  // named workflows; "main" is the entry point
	OnSucceed      []string              // shell commands to run after all stages pass
	OnFail         []string              // shell commands to run on build failure
	OnClose        []string              // shell commands to run after ticket is closed
	OnLoopComplete []string              // shell commands to run after agent loop completes
	// TemplatePromptDir is the prompts/ directory from a from: template, used as fallback for prompt resolution.
	TemplatePromptDir string

	agentExplicit bool              // true if agent: was explicitly set in config
	setFields     map[string]bool   // tracks which fields were explicitly set in config (for merge)
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

// FindConfig walks up from the tickets directory looking for .ko/config.yaml
// or .ko/pipeline.yml (legacy). Returns the path to the config file.
func FindConfig(ticketsDir string) (string, error) {
	projectRoot := ProjectRoot(ticketsDir)

	// Try new unified config.yaml first
	configPath := filepath.Join(projectRoot, ".ko", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Fall back to legacy pipeline.yml
	pipelinePath := filepath.Join(projectRoot, ".ko", "pipeline.yml")
	if _, err := os.Stat(pipelinePath); err == nil {
		return pipelinePath, nil
	}

	return "", fmt.Errorf("no config found (expected .ko/config.yaml or .ko/pipeline.yml)")
}

// FindPipelineConfig is deprecated. Use FindConfig instead.
// Kept for backwards compatibility.
func FindPipelineConfig(ticketsDir string) (string, error) {
	return FindConfig(ticketsDir)
}

// LoadConfig reads and parses a config file (.ko/config.yaml or legacy .ko/pipeline.yml).
// Returns a Config struct with both project settings and pipeline configuration.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Detect format: if we see "pipeline:" or "project:" at top level, it's the new unified format
	isUnified := false
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "pipeline:" || trimmed == "project:" {
			isUnified = true
			break
		}
	}

	if isUnified {
		return ParseConfig(content)
	}

	// Legacy format: parse as pipeline only
	pipeline, err := ParsePipeline(content)
	if err != nil {
		return nil, err
	}

	return &Config{
		Pipeline: *pipeline,
		// Project.Prefix will be empty; caller should use ReadPrefix() fallback
	}, nil
}

// LoadPipeline is deprecated. Use LoadConfig instead.
// Kept for backwards compatibility.
func LoadPipeline(path string) (*Pipeline, error) {
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}
	return &config.Pipeline, nil
}

// ParsePipeline parses pipeline YAML content (v2 format) and validates workflows.
// Uses the same minimal YAML approach as ticket parsing — no external deps.
func ParsePipeline(content string) (*Pipeline, error) {
	p, err := parsePipelineRaw(content)
	if err != nil {
		return nil, err
	}
	if err := ValidateWorkflows(p.Workflows); err != nil {
		return nil, err
	}
	return p, nil
}

// parsePipelineRaw parses pipeline YAML without validating workflows.
// Used by ParsePipeline (which adds validation) and by the from: override path
// (where workflows may be absent, inherited from the template).
func parsePipelineRaw(content string) (*Pipeline, error) {
	p := &Pipeline{
		Agent:      "claude",
		MaxRetries: 2,
		MaxDepth:   2,
		Discretion: "medium",
		Workers:    1,
		Workflows:  make(map[string]*Workflow),
		setFields:  make(map[string]bool),
	}

	lines := strings.Split(content, "\n")
	var section string        // "", "workflows", "on_succeed", "on_fail", "on_close", "on_loop_complete"
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
				p.setFields["workflows"] = true
				continue
			}
			if trimmed == "on_succeed:" {
				section = "on_succeed"
				p.setFields["on_succeed"] = true
				continue
			}
			if trimmed == "on_fail:" {
				section = "on_fail"
				p.setFields["on_fail"] = true
				continue
			}
			if trimmed == "on_close:" {
				section = "on_close"
				p.setFields["on_close"] = true
				continue
			}
			if trimmed == "on_loop_complete:" {
				section = "on_loop_complete"
				p.setFields["on_loop_complete"] = true
				continue
			}

			// Top-level scalars
			key, val, ok := parseYAMLLine(trimmed)
			if !ok {
				continue
			}
			switch key {
			case "from":
				return nil, fmt.Errorf("from: directive is only supported in unified config.yaml format (under pipeline: section)")
			case "agent":
				p.Agent = val
				p.agentExplicit = true
				p.setFields["agent"] = true
			case "command":
				p.Command = val
				p.setFields["command"] = true
			case "allow_all_tool_calls":
				p.AllowAll = val == "true"
				p.setFields["allow_all_tool_calls"] = true
			case "model":
				p.Model = val
				p.setFields["model"] = true
			case "max_retries":
				fmt.Sscanf(val, "%d", &p.MaxRetries)
				p.setFields["max_retries"] = true
			case "max_depth":
				fmt.Sscanf(val, "%d", &p.MaxDepth)
				p.setFields["max_depth"] = true
			case "discretion":
				p.Discretion = val
				p.setFields["discretion"] = true
			case "step_timeout":
				p.StepTimeout = val
				p.setFields["step_timeout"] = true
			case "require_clean_tree":
				p.RequireCleanTree = val == "true"
				p.setFields["require_clean_tree"] = true
			case "auto_triage":
				p.AutoTriage = val == "true"
				p.setFields["auto_triage"] = true
			case "auto_agent":
				p.AutoAgent = val == "true"
				p.setFields["auto_agent"] = true
			case "workers":
				fmt.Sscanf(val, "%d", &p.Workers)
				if p.Workers < 1 {
					p.Workers = 1
				}
				p.setFields["workers"] = true
			case "allowed_tools":
				p.setFields["allowed_tools"] = true
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
				case "on_success":
					currentWF.OnSuccess = val
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
		case "on_loop_complete":
			if strings.HasPrefix(trimmed, "- ") {
				cmd := strings.TrimPrefix(trimmed, "- ")
				p.OnLoopComplete = append(p.OnLoopComplete, cmd)
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

	return p, nil
}

// ParseConfig parses unified config.yaml format with project: and pipeline: sections.
func ParseConfig(content string) (*Config, error) {
	c := &Config{}

	lines := strings.Split(content, "\n")
	var section string // "project", "pipeline", or ""

	// Collect pipeline lines to parse separately
	var pipelineLines []string
	inPipeline := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and blank lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if inPipeline {
				// For blank/comment lines in pipeline, preserve them but strip 2 spaces if present
				if len(line) >= 2 && line[0] == ' ' && line[1] == ' ' {
					pipelineLines = append(pipelineLines, line[2:])
				} else {
					pipelineLines = append(pipelineLines, line)
				}
			}
			continue
		}

		// Detect top-level sections and keys (no indentation)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if trimmed == "project:" {
				section = "project"
				inPipeline = false
				continue
			}
			if trimmed == "pipeline:" {
				section = "pipeline"
				inPipeline = true
				continue
			}
			// Top-level scalar keys
			if key, val, ok := parseYAMLLine(trimmed); ok {
				inPipeline = false
				section = ""
				if idx := strings.Index(val, " #"); idx >= 0 {
					val = strings.TrimSpace(val[:idx])
				}
				switch key {
				case "summarizer":
					c.Summarizer = val
				}
				continue
			}
			// Unknown top-level key — might be pipeline content if we're in that section
			if section == "pipeline" {
				pipelineLines = append(pipelineLines, line)
				continue
			}
		}

		// Section content
		switch section {
		case "project":
			key, val, ok := parseYAMLLine(trimmed)
			if !ok {
				continue
			}
			// Strip inline comments from value
			if idx := strings.Index(val, " #"); idx >= 0 {
				val = strings.TrimSpace(val[:idx])
			}
			switch key {
			case "prefix":
				c.Project.Prefix = val
			}
		case "pipeline":
			// Strip the 2-space indentation from pipeline content
			if len(line) >= 2 && line[0] == ' ' && line[1] == ' ' {
				pipelineLines = append(pipelineLines, line[2:])
			} else {
				pipelineLines = append(pipelineLines, line)
			}
		}
	}

	// Check for from: directive in pipeline lines
	var fromDir string
	var overrideLines []string
	for _, line := range pipelineLines {
		trimmed := strings.TrimSpace(line)
		if key, val, ok := parseYAMLLine(trimmed); ok && key == "from" {
			// Strip inline comments
			if idx := strings.Index(val, " #"); idx >= 0 {
				val = strings.TrimSpace(val[:idx])
			}
			fromDir = val
			continue
		}
		overrideLines = append(overrideLines, line)
	}

	if fromDir != "" {
		expanded, err := expandTilde(fromDir)
		if err != nil {
			return nil, fmt.Errorf("pipeline from: %v", err)
		}
		if !filepath.IsAbs(expanded) {
			return nil, fmt.Errorf("pipeline from: path must be absolute or tilde-prefixed, got %q", fromDir)
		}

		base, err := LoadTemplatePipeline(expanded)
		if err != nil {
			return nil, fmt.Errorf("pipeline from: %v", err)
		}

		if len(overrideLines) > 0 {
			override, err := parsePipelineRaw(strings.Join(overrideLines, "\n"))
			if err != nil {
				return nil, fmt.Errorf("pipeline overrides: %v", err)
			}
			c.Pipeline = *MergePipeline(base, override)
		} else {
			c.Pipeline = *base
		}
		// Validate the merged result
		if err := ValidateWorkflows(c.Pipeline.Workflows); err != nil {
			return nil, err
		}
	} else if len(pipelineLines) > 0 {
		p, err := ParsePipeline(strings.Join(pipelineLines, "\n"))
		if err != nil {
			return nil, err
		}
		c.Pipeline = *p
	}

	return c, nil
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
	case "note_artifact":
		node.NoteArtifact = val
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

// expandTilde replaces a leading ~ with the user's home directory.
// Returns the path unchanged if it doesn't start with ~.
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot expand ~: %v", err)
	}
	if len(path) == 1 {
		return home, nil
	}
	// Skip ~/
	return filepath.Join(home, path[2:]), nil
}

// LoadTemplatePipeline reads and parses a template pipeline.yaml from a template
// directory. Rejects templates that themselves contain a from: directive.
func LoadTemplatePipeline(templateDir string) (*Pipeline, error) {
	pipelinePath := filepath.Join(templateDir, "pipeline.yaml")
	data, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, fmt.Errorf("template pipeline not found: %v", err)
	}

	content := string(data)

	// Guard against recursive from:
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if key, _, ok := parseYAMLLine(trimmed); ok && key == "from" {
			return nil, fmt.Errorf("template %s cannot itself contain a from: directive", pipelinePath)
		}
	}

	p, err := ParsePipeline(content)
	if err != nil {
		return nil, fmt.Errorf("parsing template pipeline %s: %v", pipelinePath, err)
	}

	p.TemplatePromptDir = filepath.Join(templateDir, "prompts")
	return p, nil
}

// MergePipeline merges an override pipeline into a base pipeline.
// Only fields explicitly set in the override (tracked via setFields) replace the base value.
// TemplatePromptDir is always preserved from the base.
func MergePipeline(base, override *Pipeline) *Pipeline {
	result := *base
	// Deep-copy maps and slices from base so we don't alias
	result.Workflows = make(map[string]*Workflow, len(base.Workflows))
	for k, v := range base.Workflows {
		result.Workflows[k] = v
	}
	if base.OnSucceed != nil {
		result.OnSucceed = append([]string(nil), base.OnSucceed...)
	}
	if base.OnFail != nil {
		result.OnFail = append([]string(nil), base.OnFail...)
	}
	if base.OnClose != nil {
		result.OnClose = append([]string(nil), base.OnClose...)
	}
	if base.OnLoopComplete != nil {
		result.OnLoopComplete = append([]string(nil), base.OnLoopComplete...)
	}
	if base.AllowedTools != nil {
		result.AllowedTools = append([]string(nil), base.AllowedTools...)
	}

	s := override.setFields
	if s["agent"] {
		result.Agent = override.Agent
		result.agentExplicit = override.agentExplicit
	}
	if s["command"] {
		result.Command = override.Command
	}
	if s["allow_all_tool_calls"] {
		result.AllowAll = override.AllowAll
	}
	if s["allowed_tools"] {
		result.AllowedTools = override.AllowedTools
	}
	if s["model"] {
		result.Model = override.Model
	}
	if s["max_retries"] {
		result.MaxRetries = override.MaxRetries
	}
	if s["max_depth"] {
		result.MaxDepth = override.MaxDepth
	}
	if s["discretion"] {
		result.Discretion = override.Discretion
	}
	if s["step_timeout"] {
		result.StepTimeout = override.StepTimeout
	}
	if s["require_clean_tree"] {
		result.RequireCleanTree = override.RequireCleanTree
	}
	if s["auto_triage"] {
		result.AutoTriage = override.AutoTriage
	}
	if s["auto_agent"] {
		result.AutoAgent = override.AutoAgent
	}
	if s["workers"] {
		result.Workers = override.Workers
	}
	if s["workflows"] {
		result.Workflows = override.Workflows
	}
	if s["on_succeed"] {
		result.OnSucceed = override.OnSucceed
	}
	if s["on_fail"] {
		result.OnFail = override.OnFail
	}
	if s["on_close"] {
		result.OnClose = override.OnClose
	}
	if s["on_loop_complete"] {
		result.OnLoopComplete = override.OnLoopComplete
	}

	// Always preserve template prompt dir from base
	result.TemplatePromptDir = base.TemplatePromptDir
	return &result
}

// LoadPromptFile reads a prompt file, searching local .ko/prompts/<name> first,
// then falling back to templatePromptDir/<name> if provided.
func LoadPromptFile(ticketsDir, name, templatePromptDir string) (string, error) {
	projectRoot := ProjectRoot(ticketsDir)
	localPath := filepath.Join(projectRoot, ".ko", "prompts", name)
	data, err := os.ReadFile(localPath)
	if err == nil {
		return string(data), nil
	}

	if templatePromptDir != "" {
		templatePath := filepath.Join(templatePromptDir, name)
		data, err = os.ReadFile(templatePath)
		if err == nil {
			return string(data), nil
		}
		return "", fmt.Errorf("prompt file '%s' not found in .ko/prompts/ or template %s", name, templatePromptDir)
	}

	return "", fmt.Errorf("prompt file '%s' not found", name)
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
