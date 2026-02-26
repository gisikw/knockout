package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectPriorContext(t *testing.T) {
	t.Run("no prior files returns empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		result := InjectPriorContext(tmpDir, "task")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("plan.md only", func(t *testing.T) {
		tmpDir := t.TempDir()
		planContent := "This is the plan"
		os.WriteFile(filepath.Join(tmpDir, "plan.md"), []byte(planContent), 0644)

		result := InjectPriorContext(tmpDir, "task")
		if result == "" {
			t.Fatal("expected non-empty result")
		}
		if !strings.Contains(result, "## Prior Context") {
			t.Error("expected '## Prior Context' header")
		}
		if !strings.Contains(result, "### plan.md") {
			t.Error("expected '### plan.md' section")
		}
		if !strings.Contains(result, planContent) {
			t.Error("expected plan content in result")
		}
	})

	t.Run("workspace task files only", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsDir := filepath.Join(tmpDir, "workspace")
		os.MkdirAll(wsDir, 0755)

		taskPlanContent := "Task plan output"
		os.WriteFile(filepath.Join(wsDir, "task.plan.md"), []byte(taskPlanContent), 0644)

		result := InjectPriorContext(tmpDir, "task")
		if result == "" {
			t.Fatal("expected non-empty result")
		}
		if !strings.Contains(result, "### task.plan.md") {
			t.Error("expected '### task.plan.md' section")
		}
		if !strings.Contains(result, taskPlanContent) {
			t.Error("expected task plan content in result")
		}
	})

	t.Run("multiple workspace files filtered by workflow", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsDir := filepath.Join(tmpDir, "workspace")
		os.MkdirAll(wsDir, 0755)

		// Create task workflow files (should be included)
		os.WriteFile(filepath.Join(wsDir, "task.plan.md"), []byte("task plan"), 0644)
		os.WriteFile(filepath.Join(wsDir, "task.implement.md"), []byte("task implement"), 0644)

		// Create bug workflow files (should be excluded)
		os.WriteFile(filepath.Join(wsDir, "bug.diagnose.md"), []byte("bug diagnose"), 0644)

		result := InjectPriorContext(tmpDir, "task")
		if result == "" {
			t.Fatal("expected non-empty result")
		}
		if !strings.Contains(result, "### task.plan.md") {
			t.Error("expected task.plan.md section")
		}
		if !strings.Contains(result, "### task.implement.md") {
			t.Error("expected task.implement.md section")
		}
		if strings.Contains(result, "bug.diagnose.md") {
			t.Error("should not include bug workflow files")
		}
	})

	t.Run("full scenario with plan and workspace files", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsDir := filepath.Join(tmpDir, "workspace")
		os.MkdirAll(wsDir, 0755)

		planContent := "Master plan"
		os.WriteFile(filepath.Join(tmpDir, "plan.md"), []byte(planContent), 0644)

		taskPlanContent := "Task planning output"
		os.WriteFile(filepath.Join(wsDir, "task.plan.md"), []byte(taskPlanContent), 0644)

		result := InjectPriorContext(tmpDir, "task")
		if result == "" {
			t.Fatal("expected non-empty result")
		}
		if !strings.Contains(result, "## Prior Context") {
			t.Error("expected header")
		}
		if !strings.Contains(result, "### plan.md") {
			t.Error("expected plan.md section")
		}
		if !strings.Contains(result, "### task.plan.md") {
			t.Error("expected task.plan.md section")
		}
		if !strings.Contains(result, planContent) {
			t.Error("expected plan content")
		}
		if !strings.Contains(result, taskPlanContent) {
			t.Error("expected task plan content")
		}
	})

	t.Run("empty files are skipped", func(t *testing.T) {
		tmpDir := t.TempDir()
		wsDir := filepath.Join(tmpDir, "workspace")
		os.MkdirAll(wsDir, 0755)

		// Create empty plan file
		os.WriteFile(filepath.Join(tmpDir, "plan.md"), []byte(""), 0644)

		// Create empty workspace file
		os.WriteFile(filepath.Join(wsDir, "task.plan.md"), []byte(""), 0644)

		result := InjectPriorContext(tmpDir, "task")
		if result != "" {
			t.Errorf("expected empty result for empty files, got %q", result)
		}
	})
}
