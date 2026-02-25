## Goal
Update README.md to reflect the simplified CLI surface after consolidation.

## Context
The CLI consolidation tickets (ko-1bfc through ko-88fa) made the following changes:
- Removed commands: `create`, `closed`, `resolved`, `add-note`, `reopen`, `query`, `questions`, `answer`, `block`, `register`, `init`, `default`, `projects`
- Updated commands:
  - `ko add` is now the only ticket creation command (removed `create` alias)
  - `ko note` is the only note command (removed `add-note` alias)
  - `ko open` is now the command to reopen (renamed from `reopen`)
  - `ko triage` now handles blocking, questions, and answers (replaced `block`, `questions`, `answer`)
  - `ko project {set,ls}` replaced `init`, `register`, `default`, `projects`
  - List commands (`ls`, `ready`, `blocked`) gained `--json` flag to replace `query`
  - `ko blocked` now accepts optional `[id]` argument to show reason for specific ticket

Current state:
- `main.go:cmdHelp()` (lines 110-155) has the correct simplified help text
- README.md still shows the old command surface:
  - Line 18: `blocked` listed without the optional `[id]` argument
  - Line 19: `closed` command still listed (removed)
  - Lines 25-26: Shows `block` and `resolved` as status commands (removed)
  - Lines 34-36: Lists `query`, `questions`, `answer` commands (removed)
  - "Plan Questions" section (lines 388-416) already correctly uses `ko triage` syntax

The README already correctly shows `ko project set` and `ko project ls` commands (lines 45-47), and the "Plan Questions" section already uses the new `ko triage` syntax.

User decision: No migration guide needed. Users can infer changes from the new command surface and error messages.

## Approach
Update the README.md command listing to match main.go's help text. Remove obsolete commands (`closed`, `block`, `resolved`, `query`, `questions`, `answer`) and update `blocked` to show it accepts an optional ID argument. Add `serve` and `triage` commands. Add `clear --force` command. The "Plan Questions" section already uses correct syntax, so no changes needed there.

## Tasks
1. [README.md:18] — Update `blocked` command description to show optional `[id]` argument matching the help text: `blocked [id]       Show blocked tickets or reason for specific ticket`
   Verify: Description matches main.go:120

2. [README.md:19] — Remove the `closed` command line entirely. This command was removed in the consolidation.
   Verify: No `closed` command in the listing

3. [README.md:21-26] — Remove the status shortcut commands `block` and `resolved` from the command listing. These are now redundant with `ko status <id> blocked|resolved`. Keep `status`, `start`, `close`, and `open`. Add the `serve` command after `open`.
   Verify: Section shows `status`, `start`, `close`, `open`, `serve` only

4. [README.md:32-36] — Add the `triage` command after `serve` section and before `dep` commands, matching main.go:128-131 format. Remove `query`, `questions`, and `answer` commands.
   Verify: `triage` command with all four forms is present, old commands are gone

5. [README.md:33] — Add `clear --force` command after `bump` command, matching main.go:139 format.
   Verify: `clear --force` command is present in the listing

## Open Questions
None. User has confirmed that no migration guide is needed.
