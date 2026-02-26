package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateWorkspace creates the workspace directory inside the artifact dir.
// Returns the workspace path.
func CreateWorkspace(artifactDir string) (string, error) {
	wsDir := filepath.Join(artifactDir, "workspace")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		return "", err
	}
	return wsDir, nil
}

// WorkspaceOutputName returns the filename for a stage output in the workspace.
// Format: <workflow>.<node>.md
func WorkspaceOutputName(workflowName, nodeName string) string {
	return workflowName + "." + nodeName + ".md"
}

// TeeOutput writes stage output to the workspace with a predictable filename.
func TeeOutput(workspaceDir, workflowName, nodeName, output string) error {
	name := WorkspaceOutputName(workflowName, nodeName)
	return os.WriteFile(filepath.Join(workspaceDir, name), []byte(output), 0644)
}

// InjectPriorContext scans the artifact directory for plan.md and workspace files
// from previous build attempts. Returns a formatted markdown string with file contents,
// or empty string if no prior context is found.
// Only includes workspace files that match the current workflow prefix (e.g., "task.*").
func InjectPriorContext(artifactDir, workflowName string) string {
	var sections []string

	// Check for plan.md at artifact root
	planPath := filepath.Join(artifactDir, "plan.md")
	if content, err := os.ReadFile(planPath); err == nil && len(content) > 0 {
		sections = append(sections, fmt.Sprintf("### plan.md\n%s", string(content)))
	}

	// Check for workspace files matching current workflow
	workspaceDir := filepath.Join(artifactDir, "workspace")
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		// Workspace directory doesn't exist or can't be read
		if len(sections) == 0 {
			return ""
		}
	} else {
		// Filter for files matching workflow prefix (e.g., "task.*.md")
		prefix := workflowName + "."
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".md") {
				continue
			}
			path := filepath.Join(workspaceDir, name)
			content, err := os.ReadFile(path)
			if err != nil || len(content) == 0 {
				continue
			}
			sections = append(sections, fmt.Sprintf("### %s\n%s", name, string(content)))
		}
	}

	if len(sections) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("## Prior Context\n\n")
	result.WriteString("From previous build attempts:\n\n")
	result.WriteString(strings.Join(sections, "\n\n"))
	return result.String()
}
