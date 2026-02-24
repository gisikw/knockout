## Goal
Enable workflows to configure whether successful builds land tickets in `resolved` or `closed` status.

## Context
The `resolved` status was added in ko-dd51 to support work that's completed but needs human review before closing. Currently, all successful builds transition tickets to `closed` (build.go:132). The pipeline needs a way to specify per-workflow whether success should land in `resolved` or `closed`.

Key files:
- `build.go`: Contains `RunBuild()` which hardcodes `setStatus(ticketsDir, t, "closed")` at line 132
- `pipeline.go`: Defines `Pipeline` and `Workflow` structs; parses pipeline.yml
- `workflow.go`: Defines `Workflow` struct (lines 40-46)
- `disposition.go`: Defines disposition types and schema
- `ticket.go`: Defines ticket statuses; `resolved` is already in the `Statuses` list (line 15)

Current workflow types in structured pipelines:
- `task`: produces code (should close on success)
- `research`: produces findings (should resolve on success for human review)
- `bug`: diagnostic workflow (should resolve on success if wontfix)

The ticket suggests both workflow-level config and a disposition override. Examining the codebase:
- Dispositions are parsed from decision node output
- Success transitions happen in `RunBuild()` after all workflows complete
- Only the terminal workflow's preference should matter (not intermediate routed workflows)

## Approach
Add a workflow-level `on_success: resolved|closed` field (defaults to `closed` for backward compatibility). Track which workflow completed last in `runWorkflow()`, then use that workflow's `on_success` config in `RunBuild()` when transitioning the ticket.

Also add a `resolved` disposition type that decision nodes can emit to override the workflow default and immediately land the ticket in `resolved` status (similar to how `fail` and `blocked` work).

## Tasks
1. [workflow.go:Workflow] — Add `OnSuccess` field (type `string`) to the `Workflow` struct.
   Verify: `go build` succeeds.

2. [pipeline.go:ParsePipeline] — Parse `on_success: resolved|closed` at workflow level in the YAML parser (around line 184 where workflow-level properties are parsed). Store in `Workflow.OnSuccess`. Default to empty string (means `closed` for backward compatibility).
   Verify: Add test case to `pipeline_test.go` and run `go test -run TestParsePipelineOnSuccess`.

3. [build.go:runWorkflow] — Change return signature from `(Outcome, error)` to include the final workflow name: `(Outcome, string, error)`. Track which workflow completed (either current workflow at the end, or the routed workflow's name from recursive call). Update all call sites.
   Verify: `go build` succeeds.

4. [build.go:RunBuild] — Update the call to `runWorkflow()` to capture the final workflow name. After success, look up that workflow's `OnSuccess` field. If it's `"resolved"`, call `setStatus(ticketsDir, t, "resolved")` instead of `"closed"`. If empty or `"closed"`, keep current behavior.
   Verify: `go test ./... -run TestRunBuild` passes (may need to update existing tests).

5. [disposition.go] — Add `"resolved"` to the `validDispositions` map (line 19). Update `DispositionSchema` constant (around line 97) to document the new disposition with an example.
   Verify: `go test -run TestParseDisposition` passes.

6. [build.go:applyDisposition] — Add a case for `"resolved"` disposition (similar to `"fail"` at line 282). Add a note to the ticket, set status to `resolved`, return `OutcomeFail` to halt the workflow (resolved tickets shouldn't continue building).
   Verify: `go test ./...` passes.

7. [pipeline_test.go] — Add test case `TestParsePipelineWorkflowOnSuccess` to verify parsing of `on_success` at workflow level (both `resolved` and `closed` values).
   Verify: `go test -run TestParsePipelineWorkflowOnSuccess` passes.

8. [examples/structured/pipeline.yml] — Add `on_success: resolved` to the `research` workflow as a demonstration.
   Verify: Visually inspect the file; no automated test needed.

## Open Questions
None. The approach is straightforward: workflow-level config for the default outcome, plus a disposition override for decision nodes that want explicit control. Backward compatibility is preserved (empty/missing `on_success` defaults to `closed`).
