**Plan summary:**

`ko triage` already exists as a list command in `cmd_list.go`. The ticket asks to extend it so that `ko triage <id> <instructions>` sets the triage field (delegating to `cmdUpdate`), following the same pattern as `cmdSnooze` and `cmdBlock`.

**Tasks:**
1. **`cmd_list.go:cmdTriage`** — After flag parsing, branch on positional arg count: 0 → list (existing), ≥2 → delegate to `cmdUpdate` with `--triage=<joined instructions>`, 1 → error.
2. **`main.go:cmdHelp`** — Update the help text for `triage` to document both forms.
3. **`cmd_list_test.go`** — Add `TestCmdTriageSet` covering set, multi-word, missing-instructions error, and 0-arg list-mode.

**No open questions** — the implementation is fully constrained by existing patterns.
