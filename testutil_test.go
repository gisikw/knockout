package main

import (
	"os"
	"testing"
)

// setupTestDB configures an isolated SQLite database for the test.
// It sets XDG_STATE_HOME to a temp directory and resets the global DB handle.
// Returns a cleanup function that should be deferred.
func setupTestDB(t *testing.T) func() {
	t.Helper()

	// Save original env
	origStateHome := os.Getenv("XDG_STATE_HOME")

	// Point to temp dir for isolated DB
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)

	// Reset global DB handle so it reinitializes with new path
	resetShadowDB()

	return func() {
		// Restore original env
		if origStateHome == "" {
			os.Unsetenv("XDG_STATE_HOME")
		} else {
			os.Setenv("XDG_STATE_HOME", origStateHome)
		}
		// Reset again so subsequent tests get fresh state
		resetShadowDB()
	}
}
