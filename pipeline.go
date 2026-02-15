package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pipeline represents a parsed .ko/pipeline.yml config.
type Pipeline struct {
	Command    string  // command to invoke for prompt stages (default: "claude")
	Model      string  // default model for prompt stages
	MaxRetries int     // max retries per stage (default: 2)
	MaxDepth   int     // max decomposition depth (default: 2)
	Discretion string  // low | medium | high (default: "medium")
	Stages     []Stage // ordered list of stages
	OnSucceed  []string // shell commands to run after all stages pass
	OnClose    []string // shell commands to run after ticket is closed
}

// Stage represents a single pipeline stage.
type Stage struct {
	Name   string // stage identifier
	Prompt string // prompt file reference (mutually exclusive with Run)
	Run    string // shell command (mutually exclusive with Prompt)
	Model  string // optional model override for this stage
	OnFail string // outcome on failure: "fail" (default) or "blocked"
}

// IsPromptStage reports whether this stage invokes an LLM.
func (s *Stage) IsPromptStage() bool {
	return s.Prompt != ""
}

// IsRunStage reports whether this stage runs a shell command.
func (s *Stage) IsRunStage() bool {
	return s.Run != ""
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

// ParsePipeline parses pipeline YAML content.
// Uses the same minimal YAML approach as ticket parsing — no external deps.
func ParsePipeline(content string) (*Pipeline, error) {
	p := &Pipeline{
		Command:    "claude",
		MaxRetries: 2,
		MaxDepth:   2,
		Discretion: "medium",
	}

	lines := strings.Split(content, "\n")
	var section string // "", "stages", "on_succeed", "on_close"
	var currentStage *Stage

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and blank lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect section headers
		if trimmed == "stages:" {
			section = "stages"
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

		// Top-level scalars (no leading whitespace)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
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
		case "stages":
			if strings.HasPrefix(trimmed, "- name:") {
				// Save previous stage
				if currentStage != nil {
					p.Stages = append(p.Stages, *currentStage)
				}
				_, val, _ := parseYAMLLine(strings.TrimPrefix(trimmed, "- "))
				currentStage = &Stage{Name: val, OnFail: "fail"}
			} else if currentStage != nil {
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				switch key {
				case "prompt":
					currentStage.Prompt = val
				case "run":
					currentStage.Run = val
				case "model":
					currentStage.Model = val
				case "on_fail":
					currentStage.OnFail = val
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

	// Save last stage
	if currentStage != nil {
		p.Stages = append(p.Stages, *currentStage)
	}

	if len(p.Stages) == 0 {
		return nil, fmt.Errorf("pipeline has no stages")
	}

	return p, nil
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
