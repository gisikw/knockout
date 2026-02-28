## Goal
Extend `cmdTriage` to accept `ko triage <id> <instructions>` as a shorthand for `ko update <id> --triage=<instructions>`, while preserving the existing no-arg list behavior.

## Context
- `cmdTriage` lives in `cmd_list.go` and currently handles only the list case (showing all tickets with a non-empty `Triage` field).
- `main.go` already routes `"triage"` to `cmdTriage(rest)`.
- The `Triage` field already exists on `Ticket` and `cmdUpdate` already handles `--triage=<value>` (see `cmd_update.go:239`).
- The analogous `cmdSnooze` pattern in `cmd_status.go:50-56` delegates to `cmdUpdate` with a constructed flag. `cmdBlock` in the same file uses `strings.Join(args[1:], " ")` to collect a multi-word reason.
- After `reorderArgs` + `resolveProjectTicketsDir` + `fs.Parse`, positional args are available via `fs.Args()` / `fs.NArg()`.

## Approach
Modify `cmdTriage` to bifurcate on the number of positional args after flag parsing: zero positional args → existing list behavior; two or more → delegate to `cmdUpdate` with `--triage=<instructions>` (joining all remaining words); exactly one → error (instructions required). Update the help text in `cmdHelp` to document the new form.

## Tasks
1. [cmd_list.go:cmdTriage] After `fs.Parse(args)` and before `ListTickets`, add a branch: if `fs.NArg() >= 2`, call `cmdUpdate([]string{fs.Arg(0), "--triage=" + strings.Join(fs.Args()[1:], " ")})` and return its result. If `fs.NArg() == 1`, print `"ko triage: instructions required"` to stderr and return 1.
   Verify: `go build ./...` succeeds.

2. [main.go:cmdHelp] Update the `triage` line in the help text to document both forms:
   `triage [<id> <instructions>]   List tickets with a triage value, or set triage on a ticket`
   Verify: `ko help` output looks right.

3. [cmd_list_test.go] Add `TestCmdTriageSet` covering: setting triage instructions saves the field, multi-word instructions are joined correctly, missing instructions (1 positional arg) returns non-zero, and calling with 0 args still performs the list (no error).
   Verify: `go test ./...` passes.

## Open Questions
None. The ticket is explicit about the desired behavior, and the existing `cmdBlock`/`cmdSnooze` patterns make the implementation unambiguous.
