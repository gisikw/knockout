The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-e8c4.artifacts/plan.md`.

## Summary

I've investigated the codebase and created a plan to update the README and help text after CLI consolidation. The key findings:

**What changed in the CLI consolidation:**
- Removed 13 commands in favor of more consolidated ones
- `ko triage` now handles blocking/questions/answers
- `ko project {set,ls}` replaced 4 separate commands
- List commands gained `--json` to replace `query`

**What needs updating:**
1. The README still shows old commands (`block`, `resolved`, `query`, `questions`, `answer`) that have been removed
2. The `blocked` command description needs updating to show it accepts an optional `[id]` argument
3. The `serve` command is missing from help text even though it exists in the code

**What's already correct:**
- `main.go` help text is up-to-date with the new CLI
- The README's "Plan Questions" section already uses the correct `ko triage` syntax
- The `project` commands are already documented correctly

I've raised one open question about whether to add a migration guide for users with scripts using the old command names.
