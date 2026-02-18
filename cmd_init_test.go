package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdInit(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	rc := cmdInit([]string{"myp"})
	if rc != 0 {
		t.Fatalf("cmdInit returned %d, want 0", rc)
	}

	// .ko/tickets/ should exist
	if info, err := os.Stat(filepath.Join(dir, ".ko", "tickets")); err != nil || !info.IsDir() {
		t.Error(".ko/tickets directory not created")
	}

	// .ko/prefix should contain "myp"
	if got := ReadPrefix(filepath.Join(dir, ".ko", "tickets")); got != "myp" {
		t.Errorf("prefix = %q, want %q", got, "myp")
	}
}

func TestCmdInitAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	// First init
	cmdInit([]string{"abc"})

	// Second init should fail
	rc := cmdInit([]string{"xyz"})
	if rc != 1 {
		t.Errorf("cmdInit on already-initialized project returned %d, want 1", rc)
	}

	// Original prefix should be preserved
	if got := ReadPrefix(filepath.Join(dir, ".ko", "tickets")); got != "abc" {
		t.Errorf("prefix = %q, want %q (should not have changed)", got, "abc")
	}
}

func TestCmdInitTooShort(t *testing.T) {
	rc := cmdInit([]string{"a"})
	if rc != 1 {
		t.Errorf("cmdInit with 1-char prefix returned %d, want 1", rc)
	}
}

func TestCmdInitNoArgs(t *testing.T) {
	rc := cmdInit(nil)
	if rc != 1 {
		t.Errorf("cmdInit with no args returned %d, want 1", rc)
	}
}
