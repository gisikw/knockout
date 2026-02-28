All tests pass. Here's a summary of what was changed:

---

**Task 1 — `cmd_list.go:cmdTriage`**

After `fs.Parse(args)` and before `ListTickets`, added a branch:
- `fs.NArg() >= 2`: delegates to `cmdUpdate` with `--triage=<joined instructions>`
- `fs.NArg() == 1`: prints `"ko triage: instructions required"` to stderr and returns 1

**Task 2 — `main.go:cmdHelp`**

Updated the `triage` help line to:
```
triage [<id> <instructions>]   List tickets with a triage value, or set triage on a ticket
```

**Task 3 — `cmd_list_test.go`**

Added `TestCmdTriageSet` with four subtests:
- `set triage saves field` — verifies single-word instructions are saved
- `multi-word instructions joined` — verifies multiple positional args are space-joined
- `missing instructions returns non-zero` — verifies 1 positional arg (ID only) returns non-zero
- `zero args performs list no error` — verifies 0 positional args runs the list path without error

No deviations from the plan.
