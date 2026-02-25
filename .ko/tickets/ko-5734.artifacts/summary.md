# Review Summary: Remove ko query command

## Compliance Check

**Plan compliance:** ✅ All planned tasks completed
- Task 1: `cmd_query.go` deleted
- Task 2: `ticketToJSON()` moved to `cmd_list.go`
- Task 3: `case "query":` removed from `main.go`
- Task 4: Help text line removed from `main.go:141`
- Task 5: `"query": true` removed from `cmd_serve.go` whitelist
- Task 6: `TestTicketQuery` deleted from `ko_test.go`
- Task 7: `testdata/ticket_query/` directory deleted
- Task 8: `specs/ticket_query.feature` deleted
- Task 9: All tests pass (`go test ./...` succeeds)

**Deviation detection:** ✅ None - implementation followed plan exactly

**Risk resolution:** ✅ No open questions in plan, no risks materialized

**Correctness:** ✅ The `ko query` command has been completely removed:
- Command no longer dispatches in main.go
- Help text no longer mentions query
- Tests no longer reference query
- Serve whitelist no longer allows query commands
- The `ticketToJSON()` helper correctly preserved in `cmd_list.go` for other commands

**Completeness:** ✅ All aspects covered:
- Source file deleted
- Tests updated to use `ko ls --json` with appropriate status filters
- Test assertions updated to match JSON array output (not JSONL)
- Spec file removed

**Safety:** ✅ No security issues:
- No secrets exposed
- No unintended deletions beyond planned removals
- Shared helper function (`ticketToJSON`) correctly preserved

**Scope:** ✅ Changes stayed within ticket scope:
- Only removed `ko query` command and related artifacts
- Updated test files to use replacement commands (`ko ls --json`)
- No unrelated changes

**Tests:** ✅ Appropriate test updates:
- Removed `TestTicketQuery` function
- Deleted test data files in `testdata/ticket_query/`
- Updated remaining tests that used `ko query` to use `ko ls --json` with status filters
- Test assertions updated for JSON array format vs JSONL
- All tests pass after changes

**Invariants:** ✅ Checked against INVARIANTS.md:
- **Specs and tests:** Spec (`specs/ticket_query.feature`) and tests (`testdata/ticket_query/`) removed together - maintains invariant that specs and tests are paired
- **Data model:** No changes to ticket structure
- **Build:** No changes to build process
- **Code organization:** `ticketToJSON()` moved to appropriate location (`cmd_list.go` where it's used)
- **Error handling:** No changes to error handling patterns
- **Naming:** N/A - removing code, not adding

## Notable Decisions

1. **Test migration strategy:** Tests that relied on `ko query` were updated to use `ko ls --json` with appropriate `--status` filters. This is the documented replacement pattern per the ticket description.

2. **JSON format mismatch:** The tests now use `ko ls --json` which outputs a JSON array, not JSONL. Test assertions were updated with regex patterns to handle whitespace variations (e.g., `"status":\s*"open"` instead of `"status":"open"`).

3. **README.md not updated:** Per ticket `ko-e8c4` (which depends on this ticket), README updates are deferred to a consolidated documentation update after all CLI consolidation work is complete. This is intentional scope management.

## Verification

- ✅ `go build` succeeds
- ✅ `go test ./...` passes (6.171s)
- ✅ `ko query` returns unknown command error
- ✅ `ko help` does not mention query command
- ✅ Serve whitelist no longer includes query
- ✅ All file deletions completed
- ✅ `ticketToJSON()` function accessible in new location

## Future Reader Notes

- The `ko query` command output all tickets as JSONL (one JSON object per line)
- Replacement: Use `ko ls --json` (outputs JSON array, all statuses), or status-specific variants with `--json` flag
- The `ticketToJSON()` helper function was preserved in `cmd_list.go` as it's used by multiple list commands for JSON serialization
- README.md still references `ko query` - this will be cleaned up by ticket `ko-e8c4`
