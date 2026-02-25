## Goal
Update README.md and main.go help text to reflect the simplified CLI surface after consolidation.

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
  - Lines 21-26: Shows `block` and `resolved` as status commands
  - Lines 34-36: Lists `query`, `questions`, `answer` commands
  - "Plan Questions" section (lines 388-416) already correctly uses `ko triage` syntax

The README already correctly shows `ko project set` and `ko project ls` commands (lines 45-47), and the "Plan Questions" section already uses the new `ko triage` syntax.

## Approach
Update the README.md command listing to match main.go's help text. Remove obsolete commands (`block`, `resolved`, `query`, `questions`, `answer`) and update `blocked` to show it accepts an optional ID argument. The "Plan Questions" section already uses correct syntax, so no changes needed there.

## Tasks
1. [README.md:18] — Update `blocked` command description to show optional `[id]` argument matching the help text: `blocked [id]       Show blocked tickets or reason for specific ticket`
   Verify: Description matches main.go:120

2. [README.md:21-26] — Remove the status shortcut commands `block` and `resolved` from the command listing. These are now redundant with `ko status <id> blocked|resolved`.
   Verify: Only `status`, `start`, `close`, and `open` remain in that section

3. [README.md:34-36] — Remove `query`, `questions`, and `answer` commands from the listing. These are replaced by `ko ls --json` and `ko triage` subcommands.
   Verify: Command list jumps directly from `bump` to `agent build`

4. [README.md:127] — Add `serve` command to the help text in main.go after the `open` command, since it exists in the command switch but is missing from help output.
   Verify: Help text includes `serve [-p port]    Start HTTP daemon (default :9876)` between `open` and the triage section

## Open Questions
Should we add a migration guide or note about the removed commands to help users who may have scripts using the old syntax? The old commands will fail with "unknown command" errors, but we could add a brief section explaining the consolidations.
