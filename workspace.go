package main

import (
	"os"
	"path/filepath"
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
