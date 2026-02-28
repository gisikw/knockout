The implementation is correct and complete. Here's my assessment:

**Plan compliance:** All three tasks completed — both `cmd_list.go` call sites updated, tests added.

**Correctness:** `&& t.Triage == ""` added at both filter sites (local queue line 230, cross-project queue line 276), exactly as specified.

**Tests:** 
- `TestReadyTriageExclusion` in `ticket_test.go` documents the predicate
- `testdata/ticket_triage/ready_exclusion.txtar` is a real integration test mirroring the snooze pattern, picked up automatically by `TestTicketTriage`

**Spec:** Two new scenarios in `specs/ticket_triage.feature` match the INVARIANTS.md requirement that every behavior has a spec.

**INVARIANTS.md:** No violations. Spec and test are both present. The inline check is consistent with the "decision logic is pure" invariant (no logic mixed with I/O).

**Tests:** `go test ./...` passes cleanly.

```json
{"disposition": "continue"}
```
