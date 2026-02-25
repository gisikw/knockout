package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed agent-harnesses/*.sh
var embeddedHarnesses embed.FS

// HarnessConfig represents a loaded shell harness.
type HarnessConfig struct {
	ScriptPath string
}

// LoadHarness loads a shell harness by name from project config, user config, or built-ins.
// Search order: .ko/agent-harnesses/ → ~/.config/knockout/agent-harnesses/ → embedded built-ins
func LoadHarness(name string) (*HarnessConfig, error) {
	// Try project-local config
	projectShellPath := filepath.Join(".ko", "agent-harnesses", name)
	if info, err := os.Stat(projectShellPath); err == nil && info.Mode()&0111 != 0 {
		return &HarnessConfig{
			ScriptPath: projectShellPath,
		}, nil
	}

	// Try user config
	if home, err := os.UserHomeDir(); err == nil {
		userShellPath := filepath.Join(home, ".config", "knockout", "agent-harnesses", name)
		if info, err := os.Stat(userShellPath); err == nil && info.Mode()&0111 != 0 {
			return &HarnessConfig{
				ScriptPath: userShellPath,
			}, nil
		}
	}

	// Try embedded built-ins
	embeddedShellPath := filepath.Join("agent-harnesses", name+".sh")
	if data, err := embeddedHarnesses.ReadFile(embeddedShellPath); err == nil {
		// Write to temp location with exec permissions
		tmpFile, err := os.CreateTemp("", "ko-harness-"+name+"-*.sh")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for harness: %w", err)
		}
		if err := os.WriteFile(tmpFile.Name(), data, 0755); err != nil {
			return nil, fmt.Errorf("failed to write harness script: %w", err)
		}
		return &HarnessConfig{
			ScriptPath: tmpFile.Name(),
		}, nil
	}

	return nil, fmt.Errorf("harness %q not found", name)
}


// ShellAdapter implements AgentAdapter by executing a shell script with KO_* env vars.
type ShellAdapter struct {
	scriptPath string
}

// NewShellAdapter creates a ShellAdapter from a script path.
func NewShellAdapter(scriptPath string) *ShellAdapter {
	return &ShellAdapter{scriptPath: scriptPath}
}

// BuildCommand sets KO_* environment variables and executes the shell script.
func (a *ShellAdapter) BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd {
	cmd := exec.Command(a.scriptPath)

	// Set KO_* environment variables
	cmd.Env = append(os.Environ(),
		"KO_PROMPT="+prompt,
		"KO_MODEL="+model,
		"KO_SYSTEM_PROMPT="+systemPrompt,
	)

	// Set KO_ALLOW_ALL as "true" or "false"
	if allowAll {
		cmd.Env = append(cmd.Env, "KO_ALLOW_ALL=true")
	} else {
		cmd.Env = append(cmd.Env, "KO_ALLOW_ALL=false")
	}

	// Set KO_ALLOWED_TOOLS as comma-separated string
	if len(allowedTools) > 0 {
		toolsCSV := strings.Join(allowedTools, ",")
		cmd.Env = append(cmd.Env, "KO_ALLOWED_TOOLS="+toolsCSV)
	} else {
		cmd.Env = append(cmd.Env, "KO_ALLOWED_TOOLS=")
	}

	return cmd
}
