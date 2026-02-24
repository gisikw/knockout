## Goal
Add configurable step timeout to pipeline nodes to prevent indefinite hangs with a 15-minute default.

## Context
Pipeline execution happens in `build.go` via `RunBuild` → `runWorkflow` → `runNode`. Individual nodes invoke either:
- `runPromptNode` (build.go:346) for LLM-based nodes using `exec.Command`
- `runRunNode` (build.go:409) for shell command nodes using `exec.Command`

Neither currently uses `exec.CommandContext`, so commands can run indefinitely. The pipeline config is parsed in `pipeline.go:ParsePipeline` which reads YAML and populates a `Pipeline` struct. Node definitions are in `workflow.go:Node`. The codebase uses `time.ParseDuration` (see cmd_loop.go:82) for parsing duration strings like "15m".

Existing test patterns use testscript in `testdata/*.txtar` files with fake-llm executables to simulate agent behavior. Specs live in `specs/*.feature` using Gherkin syntax.

INVARIANTS.md requires:
- Specs before code
- Pure decision functions (no I/O in decision logic)
- 500-line file limit
- Tests mirror source files

## Approach
1. Add `StepTimeout` field to `Pipeline` and `Timeout` field to `Node` structs
2. Parse these fields in `ParsePipeline` using Go's duration format (e.g., "15m")
3. Add a `parseTimeout` helper to convert duration strings to `time.Duration`
4. Modify `runPromptNode` and `runRunNode` to use `exec.CommandContext` with a deadline derived from the timeout (node timeout → workflow timeout → pipeline default → 15m hardcoded default)
5. Handle timeout errors explicitly with clear error messages
6. Add timeout field to event logging for diagnosability

## Tasks

1. [specs/pipeline.feature] — Add scenarios for step timeout behavior.
   - Scenario: Pipeline-level step_timeout sets default
   - Scenario: Per-node timeout overrides the pipeline default
   - Scenario: Default timeout is 15 minutes when not specified
   - Scenario: Timed-out steps fail with clear error message
   - Scenario: Both prompt nodes and run nodes respect timeout
   Verify: Spec is clear and covers all acceptance criteria.

2. [workflow.go:Node] — Add `Timeout` field to Node struct (type `string`).
   Verify: Field is exported and documented.

3. [pipeline.go:Pipeline] — Add `StepTimeout` field to Pipeline struct (type `string`).
   Verify: Field is exported and documented.

4. [pipeline.go:ParsePipeline] — Parse `step_timeout` at pipeline level and `timeout` at node level from YAML.
   Apply to appropriate structs when found. Use the same pattern as existing scalar parsing (e.g., "model", "max_retries").
   Verify: `go test ./...` passes.

5. [pipeline.go] — Add `parseTimeout(durationStr string) (time.Duration, error)` function.
   Returns parsed duration using `time.ParseDuration`. Returns 15*time.Minute if input is empty.
   Verify: Pure function, easily unit testable.

6. [pipeline_test.go] — Add unit tests for `parseTimeout`.
   Test cases: empty string returns 15m, "5m" returns 5 minutes, "1h30m" returns 90 minutes, invalid format returns error.
   Verify: Tests pass.

7. [build.go:runNode] — Add timeout resolution logic: resolve the effective timeout in precedence order (node → pipeline → 15m default).
   Parse the timeout string using `parseTimeout`. Pass the resolved `time.Duration` to `runPromptNode` and `runRunNode`.
   Verify: Logic follows same pattern as `resolveModel` and `resolveAllowAll`.

8. [build.go:runPromptNode] — Add `timeout time.Duration` parameter.
   Replace `cmd := adapter.BuildCommand(...)` with `ctx, cancel := context.WithTimeout(context.Background(), timeout); defer cancel()`.
   Replace `cmd.Output()` with `cmd.CommandContext(ctx, ...)` followed by `cmd.Output()`.
   Check if error is `context.DeadlineExceeded` and return a clear timeout error message.
   Verify: Timeout kills the process and reports "step timed out after Xm".

9. [build.go:runRunNode] — Add `timeout time.Duration` parameter.
   Replace `exec.Command("sh", "-c", node.Run)` with context-based command.
   Use `ctx, cancel := context.WithTimeout(context.Background(), timeout); defer cancel()`.
   Create command with `exec.CommandContext(ctx, "sh", "-c", node.Run)`.
   Check if error is `context.DeadlineExceeded` and return clear timeout error.
   Verify: Shell commands respect timeout and report clear errors on expiry.

10. [build.go:runCmdVerbose] — Update signature to accept `*exec.Cmd` that may already have a context.
    Ensure it respects the context from the caller (don't replace it).
    Verify: Verbose mode works with timeouts.

11. [testdata/pipeline/step_timeout_default.txtar] — Test that default 15-minute timeout is applied when not specified.
    Use a fake-llm that sleeps and verify the build doesn't hang indefinitely (test with a short timeout like 2s to avoid slow tests).
    Verify: Test passes, demonstrates default behavior.

12. [testdata/pipeline/step_timeout_node_override.txtar] — Test that node-level timeout overrides pipeline default.
    Set pipeline default to 10s, node override to 2s, verify node timeout is used.
    Verify: Test passes.

13. [testdata/pipeline/step_timeout_message.txtar] — Test that timeout error message is clear.
    Set short timeout, verify error contains "timed out after" and duration.
    Verify: Test passes, error is diagnosable.

14. [testdata/pipeline/step_timeout_run_node.txtar] — Test that run: nodes (shell commands) also respect timeout.
    Create a run node that sleeps longer than timeout, verify it's killed.
    Verify: Test passes, shell commands time out correctly.

15. [pipeline_test.go] — Add integration test for timeout parsing in pipeline config.
    Test that `step_timeout: 20m` in YAML results in correct Pipeline.StepTimeout value.
    Test that `timeout: 5m` in node YAML results in correct Node.Timeout value.
    Verify: Tests pass.

## Open Questions

**Q: Should the timeout be enforced in verbose mode differently, or should verbose output also be killed on timeout?**
A: Verbose mode should respect timeouts the same way. The context will kill the process regardless of output streaming.

**Q: Should we log the timeout value at the start of each node execution?**
A: Yes, add it to the NodeStart event log for diagnosability. This will help understand why a step was killed.

**Q: Do we need to handle Workflow-level timeout (in addition to pipeline and node)?**
A: The ticket design doesn't mention workflow-level timeout, and the pattern of override is simpler with just pipeline → node. Not implementing workflow-level override keeps it simple.

**Q: Should timeout be mandatory in the YAML or optional with a safe default?**
A: Optional with a 15-minute default per the ticket design. This prevents breaking existing pipeline configs.
