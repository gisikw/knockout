## Goal
Add support for post-loop hooks that run once after the agent loop completes, separate from per-ticket `on_succeed`/`on_fail`/`on_close` hooks.

## Context
Currently, the pipeline system has three hook types that run per-ticket:
- `on_succeed` — runs after workflow succeeds, before ticket closes
- `on_fail` — runs when a build fails (best-effort cleanup)
- `on_close` — runs after ticket is closed (safe for deploys)

The agent loop (`ko agent loop`) burns down the ready queue, building tickets one at a time until the queue is empty or limits are reached. The loop stops for one of these reasons:
- `empty` — ready queue is exhausted
- `max_tickets` — ticket limit reached
- `max_duration` — time limit reached
- `build_error` — fatal build execution error

**Key files:**
- `pipeline.go` — Pipeline struct definition (lines 24-40), currently has OnSucceed, OnFail, OnClose
- `loop.go` — RunLoop function (lines 66-139), main loop logic
- `cmd_loop.go` — cmdAgentLoop function (lines 83-221), CLI command handler
- `build.go` — RunBuild function (lines 58-163), per-ticket build execution

**Current behavior:**
The loop writes a JSONL summary to `.ko/agent.log` after completion (cmd_loop.go:215) but has no mechanism for running arbitrary shell commands after the loop finishes.

**Architectural context (from INVARIANTS.md):**
- "Every behavior has a spec" — need to add a spec to `specs/loop.feature`
- "Every spec has a test" — need a txtar test in `testdata/loop/`
- The pipeline config uses YAML parsing without external dependencies
- Hooks are shell commands with env vars like `$TICKET_ID`, `$CHANGED_FILES`, `$KO_TICKET_WORKSPACE`

**Ticket request:**
Add hooks that run once when the loop completes, not per-ticket. Use case: updating README, pushing changes after all tickets are processed.

## Approach
Add an `on_loop_complete` hook field to the Pipeline struct that runs shell commands after the agent loop finishes, regardless of stop reason. These hooks will have access to loop summary statistics via environment variables so they can make intelligent decisions (e.g., only push if at least one ticket succeeded).

The hook runs after the loop completes in all exit paths except panics, similar to how per-ticket hooks work. Hook failures are logged but don't affect the loop exit code or results.

## Tasks
1. [pipeline.go:Pipeline] — Add `OnLoopComplete []string` field to Pipeline struct after OnClose field (around line 37). Add corresponding comment explaining these run after loop completes.
   Verify: `go build ./...` succeeds.

2. [pipeline.go:ParsePipeline] — Add parsing logic for `on_loop_complete:` section in the YAML parser, similar to on_succeed/on_fail/on_close sections (around lines 357-361). Add case in switch statement to handle "on_loop_complete" section that appends hook commands to `p.OnLoopComplete`.
   Verify: `go test ./... -run TestParse` passes.

3. [loop.go] — Add `runLoopHooks` helper function similar to `runHooks` in build.go but accepting LoopResult and elapsed duration as parameters instead of Ticket. Set env vars: `LOOP_PROCESSED`, `LOOP_SUCCEEDED`, `LOOP_FAILED`, `LOOP_BLOCKED`, `LOOP_DECOMPOSED`, `LOOP_STOPPED`, `LOOP_RUNTIME_SECONDS`.
   Verify: `go build ./...` succeeds.

4. [cmd_loop.go:cmdAgentLoop] — Call `runLoopHooks` after line 215 (after writeAgentLogSummary) but before cleanup. Pass ticketsDir, pipeline, result, and elapsed. Log any errors to stderr but don't change exit code.
   Verify: `go build ./...` succeeds.

5. [specs/loop.feature] — Add new scenario: "on_loop_complete hooks run after loop finishes". Specify that hooks have access to loop summary env vars and run regardless of stop reason.
   Verify: Spec follows existing Gherkin style in file.

6. [testdata/loop/loop_on_complete_hook.txtar] — Create txtar test with pipeline config containing on_loop_complete hook that writes env vars to a file. Verify hook runs and env vars are set correctly.
   Verify: `go test ./... -run TestScript/loop_on_complete_hook` passes.

7. [examples/default/pipeline.yml] — Add commented-out example of on_loop_complete hooks with explanation (around line 38, after on_close example).
   Verify: Example uses correct YAML syntax and clear comment.

8. [README.md] — Document on_loop_complete hook in the "Hooks" section (around lines 307-314). Explain available env vars and when hooks run.
   Verify: Documentation is consistent with existing hook documentation style.

## Open Questions
None — the implementation is straightforward and follows existing patterns for hooks. The only design choice is which environment variables to expose, and the loop result struct already provides all relevant information (Processed, Succeeded, Failed, Blocked, Decomposed, Stopped, runtime).
