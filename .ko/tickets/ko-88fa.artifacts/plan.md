## Goal
Update the ko agent build pipeline and HTTP server to use the new CLI command syntax after the CLI simplification.

## Context
The ticket references a CLI simplification that renamed several commands:
- `create` → `add`
- `add-note` → `note`
- `reopen` → `open`
- `blocked`/`closed`/`resolved` → `ls --status=X`
- `query` → `ls --json`
- `block`/`questions`/`answer` → `triage`
- `init`/`register`/`default` → `project set`
- `projects` → `project ls`

After searching the codebase, I found:

**cmd_serve.go (lines 325-343)**: The HTTP `/ko` endpoint has a whitelist of allowed subcommands. Currently includes `"query"` (line 331 in cmd_serve.go, lines 31, 204, 213 in cmd_serve_test.go) which is the old command name that was replaced by `ls --json`. However, the whitelist correctly does NOT include `create`, `add-note`, `reopen`, `init`, `register`, or `projects`.

**cmd_serve_test.go (line 238)**: A test comment explicitly lists `"create", "add", "init"` as excluded commands, confirming these old names should not be in the whitelist. The test also checks that `"query"` is in the whitelist, which needs updating.

**Prompt files**: None of the `.ko/prompts/*.md` files contain references to old command names. They reference generic concepts like "create tickets" but don't hardcode specific CLI command invocations.

**build.go**: The build pipeline execution uses `exec.Command` for hooks and node commands, but doesn't hardcode any ko CLI commands internally. Hooks are user-defined in `.ko/config.yaml`.

**main.go**: Command dispatch is correct - `add` maps to `cmdCreate`, `note` maps to `cmdAddNote`, etc. The CLI interface itself is already updated.

## Approach
The primary issue is the HTTP server whitelist in `cmd_serve.go` which still includes the old `query` command. This needs to be removed since the new syntax is `ls --json`. The whitelist should only allow commands that are safe for remote execution, and `query` no longer exists as a standalone command.

## Tasks
1. [cmd_serve.go:326-343] — Remove `"query": true` from the whitelist map in `cmdServe` function.
   Verify: The whitelist no longer contains `query`.

2. [cmd_serve_test.go:27-44] — Remove `"query": true` from the test whitelist map (around line 31).
   Verify: Test data matches production code.

3. [cmd_serve_test.go:200-246] — Update `TestServeWhitelist` expected commands list to remove `"query"` from line 204, and update the corresponding whitelist map to remove `"query": true` from around line 213.
   Verify: `go test -run TestServeWhitelist` passes.

4. [cmd_serve_test.go:238] — Update the test comment to clarify that `"query"` is also explicitly excluded (old command name). Change comment from `"create", "add", "init", // explicitly excluded per plan` to `"create", "add", "init", "query", // old command names excluded`.
   Verify: Comment accurately reflects the architectural decision.

5. Run full test suite to ensure no other tests depend on `query` in the whitelist.
   Verify: `go test ./...` passes.

## Open Questions
None. The change is straightforward: remove the obsolete `query` command from the HTTP server whitelist. All other old command names are already absent from the whitelist. The new command `ls --json` is already supported via the existing `"ls"` whitelist entry (clients can pass `["ls", "--json"]` in the argv).
