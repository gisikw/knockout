package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func cmdAgentInit(args []string) int {
	// Resolve project from #tag or fall back to local project root
	var projectRoot string
	ticketsDir, _, err := resolveProjectTicketsDir(args)
	if err == nil {
		projectRoot = ProjectRoot(ticketsDir)
	} else {
		projectRoot, err = findProjectRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko agent init: %v\n", err)
			return 1
		}
	}

	koDir := filepath.Join(projectRoot, ".ko")
	configPath := filepath.Join(koDir, "pipeline.yml")

	// Don't overwrite existing config
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(os.Stderr, "ko agent init: %s already exists\n", configPath)
		return 1
	}

	promptsDir := filepath.Join(koDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent init: %v\n", err)
		return 1
	}

	// Ensure builds dir is gitignored
	buildsDir := filepath.Join(koDir, "builds")
	if err := os.MkdirAll(buildsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent init: %v\n", err)
		return 1
	}
	os.WriteFile(filepath.Join(buildsDir, ".gitignore"), []byte("*\n!.gitignore\n"), 0644)

	// Gitignore agent runtime files in .ko/
	os.WriteFile(filepath.Join(koDir, ".gitignore"), []byte("agent.lock\nagent.pid\nagent.log\nagent.heartbeat\n"), 0644)

	// Write pipeline config
	if err := os.WriteFile(configPath, []byte(defaultPipelineYML), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent init: %v\n", err)
		return 1
	}

	// Write prompt files
	prompts := map[string]string{
		"triage.md":    defaultTriagePrompt,
		"implement.md": defaultImplementPrompt,
		"review.md":    defaultReviewPrompt,
	}
	for name, content := range prompts {
		path := filepath.Join(promptsDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "ko agent init: %v\n", err)
			return 1
		}
	}

	fmt.Println("initialized .ko/pipeline.yml")
	fmt.Println("  .ko/prompts/triage.md")
	fmt.Println("  .ko/prompts/implement.md")
	fmt.Println("  .ko/prompts/review.md")
	return 0
}

const defaultPipelineYML = `# Build pipeline -- triage, implement, verify, review
#
# Triage is a decision node: it evaluates the ticket and either continues
# to implementation or signals fail/blocked/decompose via a JSON disposition.
#
# Implement and verify are action nodes: they just do work, no output parsing.
#
# Review is a decision node: it evaluates the implementation and either
# continues (succeed) or signals fail.
model: claude-sonnet-4-5-20250929
max_retries: 2
max_depth: 2
discretion: medium

workflows:
  main:
    - name: triage
      type: decision
      prompt: triage.md
    - name: implement
      type: action
      prompt: implement.md
    - name: verify
      type: action
      run: just test
    - name: review
      type: decision
      prompt: review.md

# Commands to run after all nodes succeed (before ticket is closed).
# $TICKET_ID, $CHANGED_FILES, and $KO_TICKET_WORKSPACE are available.
# on_succeed:
#   - git add -A
#   - git commit -m "ko: implement ${TICKET_ID}"

# Commands to run on build failure (cleanup worktree, reset state, etc.).
# on_fail:
#   - git checkout -- .

# Commands to run after ticket is closed (safe for deploys).
# on_close:
#   - git push
`

const defaultTriagePrompt = `You are triaging a ticket to determine if it's ready for automated implementation.

**Before concluding anything, investigate the codebase.** Search for relevant
code, read the files involved, and understand the current implementation. Many
tickets are terse but perfectly actionable once you see the code they refer to.

Evaluate the ticket:

1. **Is the scope clear?** Can you identify exactly what needs to change?
   Search the codebase for relevant strings, types, or patterns mentioned in
   the ticket.
2. **Are the files identifiable?** Use grep/glob to find them.
3. **Is it self-contained?** Can this be done without human decisions?
4. **Are there acceptance criteria?** Either explicit or clearly implied from
   the current code and the requested change?

If the ticket is actionable, provide:
- A brief summary of what needs to be done
- The files you expect to modify
- Any assumptions you're making

Then end with a ` + "`continue`" + ` disposition.

If the ticket is genuinely ambiguous *after* you've looked at the code, end
with a ` + "`fail`" + ` disposition explaining what's missing.

If implementing this ticket requires something else to be done first that isn't
captured in the ticket's dependencies, end with a ` + "`blocked`" + ` disposition
identifying the blocker.

If the ticket is too large for a single implementation pass, end with a
` + "`decompose`" + ` disposition listing the subtasks.
`

const defaultImplementPrompt = `You are implementing a ticket. The triage stage has already confirmed this ticket
is actionable, and its analysis is provided as previous stage output.

Implement the changes described in the ticket. Follow these rules:

1. **Read before writing.** Always read existing files before modifying them.
2. **Read INVARIANTS.md** (if it exists in the project root) before writing any
   code. These are architectural contracts -- your implementation must comply.
3. **Minimal changes.** Only change what the ticket requires. Don't refactor
   surrounding code, add comments to unchanged code, or "improve" things that
   aren't broken.
4. **Follow existing patterns.** Match the style, naming conventions, and
   architecture of the existing codebase.
5. **No new dependencies** unless the ticket explicitly calls for them.
6. **Write tests** if the codebase has tests and the change is testable.
7. **Do NOT commit, push, or close the ticket.** Leave changes uncommitted.
   The pipeline handles git operations and ticket lifecycle separately.

When you're done, provide a brief summary of what you changed and why.
`

const defaultReviewPrompt = `You are reviewing changes made by an automated implementation stage.

Look at the git diff of uncommitted changes and evaluate:

1. **Correctness.** Does the implementation match what the ticket asked for?
2. **Completeness.** Is anything missing? Are edge cases handled?
3. **Safety.** Any security issues (injection, XSS, leaked secrets)?
   Any accidental deletions or unintended side effects?
4. **Scope.** Did the implementation stay within the ticket's scope, or did it
   make unrelated changes?
5. **Tests.** If the codebase has tests, were appropriate tests added/updated?
6. **Invariants.** Read INVARIANTS.md (if it exists in the project root) and
   check the diff against every documented contract. These are non-negotiable --
   a violation is a blocker.

If the changes look good, end with a ` + "`continue`" + ` disposition.

If there are problems that you can fix, fix them and end with ` + "`continue`" + `.

If the problems indicate a fundamental misunderstanding of the ticket or an
invariant violation that you cannot fix, end with a ` + "`fail`" + ` disposition
explaining what went wrong.
`
