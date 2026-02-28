All tasks from the plan were completed correctly:

- `IsSnoozed` pure function added to `ticket.go` with correct midnight UTC semantics
- `TestIsSnoozed` with 5 deterministic cases (empty, past, future, today boundary, invalid)
- All three `IsReady` call sites updated in `cmd_list.go` and `loop.go`
- Spec scenarios added to `ticket_snooze.feature`
- `testdata/ticket_snooze/ready_exclusion.txtar` integration test created and passing
- No invariant violations; build and tests pass

```json
{"disposition": "continue"}
```
