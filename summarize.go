package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TitleMaxWords is the threshold above which a title gets auto-summarized.
const TitleMaxWords = 12

// SummarizeTitle runs the summarizer command to condense a long title.
// Returns the summarized title or the original if summarization fails.
func SummarizeTitle(summarizer, title string) string {
	prompt := fmt.Sprintf(
		"Summarize the following in 8 words or fewer. Output ONLY the summary, nothing else.\n\n%s",
		title,
	)

	// Use _daemonize to break out of any nested Claude session.
	// The summarizer command might be claude, which would fail inside
	// a Claude session without the double-fork env stripping.
	self, err := os.Executable()
	if err != nil {
		return title
	}

	// Build: ko agent _summarize <summarizer-cmd> with prompt on stdin
	cmd := exec.Command(self, "agent", "_summarize", summarizer)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: summarizer failed: %v\n", err)
		return title
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return title
	}
	// Strip surrounding quotes that LLMs like to add
	result = strings.Trim(result, "\"'`")
	return result
}
