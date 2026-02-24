package main

import (
	"testing"
	"time"
)

func TestShouldContinueUnlimited(t *testing.T) {
	c := LoopConfig{}
	ok, _ := c.ShouldContinue(100, time.Hour)
	if !ok {
		t.Error("unlimited config should always continue")
	}
}

func TestShouldContinueMaxTickets(t *testing.T) {
	c := LoopConfig{MaxTickets: 3}

	ok, _ := c.ShouldContinue(2, 0)
	if !ok {
		t.Error("should continue when under limit")
	}

	ok, reason := c.ShouldContinue(3, 0)
	if ok {
		t.Error("should stop at limit")
	}
	if reason != "max_tickets" {
		t.Errorf("reason = %q, want %q", reason, "max_tickets")
	}
}

func TestShouldContinueMaxDuration(t *testing.T) {
	c := LoopConfig{MaxDuration: 5 * time.Minute}

	ok, _ := c.ShouldContinue(0, 4*time.Minute)
	if !ok {
		t.Error("should continue when under duration")
	}

	ok, reason := c.ShouldContinue(0, 5*time.Minute)
	if ok {
		t.Error("should stop at duration limit")
	}
	if reason != "max_duration" {
		t.Errorf("reason = %q, want %q", reason, "max_duration")
	}
}

func TestShouldContinueMaxTicketsTakesPrecedence(t *testing.T) {
	c := LoopConfig{MaxTickets: 1, MaxDuration: time.Hour}

	ok, reason := c.ShouldContinue(1, 30*time.Minute)
	if ok {
		t.Error("should stop when ticket limit reached")
	}
	if reason != "max_tickets" {
		t.Errorf("reason = %q, want %q", reason, "max_tickets")
	}
}

func TestLoopResult(t *testing.T) {
	result := LoopResult{
		Processed:  10,
		Succeeded:  5,
		Failed:     2,
		Blocked:    2,
		Decomposed: 1,
		Stopped:    "max_tickets",
	}

	// Verify all fields are accessible and have expected values
	if result.Processed != 10 {
		t.Errorf("Processed = %d, want 10", result.Processed)
	}
	if result.Succeeded != 5 {
		t.Errorf("Succeeded = %d, want 5", result.Succeeded)
	}
	if result.Failed != 2 {
		t.Errorf("Failed = %d, want 2", result.Failed)
	}
	if result.Blocked != 2 {
		t.Errorf("Blocked = %d, want 2", result.Blocked)
	}
	if result.Decomposed != 1 {
		t.Errorf("Decomposed = %d, want 1", result.Decomposed)
	}
	if result.Stopped != "max_tickets" {
		t.Errorf("Stopped = %q, want %q", result.Stopped, "max_tickets")
	}
}
