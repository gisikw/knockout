package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed agent-harnesses/*.yaml
var embeddedHarnesses embed.FS

// Harness describes how to build a command for an agent.
type Harness struct {
	Binary          string   `yaml:"binary"`
	BinaryFallbacks []string `yaml:"binary_fallbacks"`
	Args            []string `yaml:"args"`
}

// LoadHarness loads a harness by name from project config, user config, or built-ins.
// Search order: .ko/agent-harnesses/ → ~/.config/knockout/agent-harnesses/ → embedded built-ins
func LoadHarness(name string) (*Harness, error) {
	filename := name + ".yaml"

	// Try project-local config
	projectPath := filepath.Join(".ko", "agent-harnesses", filename)
	if data, err := os.ReadFile(projectPath); err == nil {
		return parseHarness(data)
	}

	// Try user config
	if home, err := os.UserHomeDir(); err == nil {
		userPath := filepath.Join(home, ".config", "knockout", "agent-harnesses", filename)
		if data, err := os.ReadFile(userPath); err == nil {
			return parseHarness(data)
		}
	}

	// Try embedded built-ins
	embeddedPath := filepath.Join("agent-harnesses", filename)
	if data, err := embeddedHarnesses.ReadFile(embeddedPath); err == nil {
		return parseHarness(data)
	}

	return nil, fmt.Errorf("harness %q not found", name)
}

func parseHarness(data []byte) (*Harness, error) {
	var h Harness
	if err := yaml.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("invalid harness YAML: %w", err)
	}
	return &h, nil
}

// TemplateAdapter implements AgentAdapter by rendering a harness template.
type TemplateAdapter struct {
	harness *Harness
}

// NewTemplateAdapter creates a TemplateAdapter from a harness.
func NewTemplateAdapter(h *Harness) *TemplateAdapter {
	return &TemplateAdapter{harness: h}
}

// BuildCommand renders the harness template and returns a ready-to-run Cmd.
func (a *TemplateAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd {
	// Resolve binary
	bin := a.resolveBinary()

	// Build template variables
	vars := map[string]string{
		"prompt":             prompt,
		"model":              model,
		"system_prompt":      systemPrompt,
		"prompt_with_system": buildPromptWithSystem(prompt, systemPrompt),
		"allow_all":          "",
		"cursor_allow_all":   "",
		"allowed_tools":      "",
	}

	// Set conditional flags
	if model != "" {
		vars["model"] = "--model\n" + model
	}
	if systemPrompt != "" {
		vars["system_prompt"] = "--append-system-prompt\n" + systemPrompt
	}
	if allowAll {
		vars["allow_all"] = "--dangerously-skip-permissions"
		vars["cursor_allow_all"] = "--force"
	}
	if len(allowedTools) > 0 {
		toolsCSV := strings.Join(allowedTools, ",")
		vars["allowed_tools"] = "--allowed-prompts\n" + toolsCSV
	}

	// Render args
	var renderedArgs []string
	useStdin := false
	for i, arg := range a.harness.Args {
		// Check if this is the stdin marker (standalone "-p" without template variable following)
		if arg == "-p" && (i+1 >= len(a.harness.Args) || !strings.Contains(a.harness.Args[i+1], "${prompt")) {
			renderedArgs = append(renderedArgs, arg)
			useStdin = true
			continue
		}

		// Render template variables
		rendered := arg
		for key, val := range vars {
			rendered = strings.ReplaceAll(rendered, "${"+key+"}", val)
		}

		// If this was a pure template variable for flags (model, system_prompt, allow_all),
		// split on newlines to allow conditional multi-arg expansion.
		// Don't split prompt content (prompt, prompt_with_system) which may contain newlines.
		isPromptContent := arg == "${prompt}" || arg == "${prompt_with_system}"
		if strings.HasPrefix(arg, "${") && strings.HasSuffix(arg, "}") && !isPromptContent {
			// Pure template variable for flags - split on newlines for multi-arg expansion
			parts := strings.Split(rendered, "\n")
			for _, part := range parts {
				if part != "" {
					renderedArgs = append(renderedArgs, part)
				}
			}
		} else {
			// Prompt content, mixed content, or literal - keep as-is
			if rendered != "" {
				renderedArgs = append(renderedArgs, rendered)
			}
		}
	}

	cmd := exec.Command(bin, renderedArgs...)

	// Set stdin if using -p flag for stdin
	if useStdin {
		cmd.Stdin = strings.NewReader(prompt)
	}

	return cmd
}

// resolveBinary resolves the binary path, trying fallbacks if defined.
func (a *TemplateAdapter) resolveBinary() string {
	// If binary_fallbacks is defined, try each in order
	if len(a.harness.BinaryFallbacks) > 0 {
		for _, bin := range a.harness.BinaryFallbacks {
			if path, err := exec.LookPath(bin); err == nil {
				return path
			}
		}
		// Fall back to first option (will fail at exec time with clear error)
		return a.harness.BinaryFallbacks[0]
	}

	// Otherwise use the binary field directly
	return a.harness.Binary
}

// buildPromptWithSystem combines system prompt and user prompt for agents that don't support separate system prompts.
func buildPromptWithSystem(prompt, systemPrompt string) string {
	if systemPrompt == "" {
		return prompt
	}
	return systemPrompt + "\n\n" + prompt
}
