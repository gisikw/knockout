# Implementation Summary: Add --all flag to ko ls

## What Was Done

Successfully implemented the `--all` flag for `ko ls` to provide parity with the SSE behavior. The flag allows users to include closed tickets in the listing output.

### Changes Made

1. **cmd_list.go** (lines 139, 159-160, 179-180)
   - Added `allTickets := fs.Bool("all", false, "include closed tickets")` flag
   - Modified filtering logic in both JSON and non-JSON output paths to check `!*allTickets` before applying the default closed-ticket filter
   - Maintained backward compatibility: default behavior (without `--all`) excludes closed tickets

2. **specs/ticket_listing.feature** (lines 26-33)
   - Added "List with --all flag includes closed tickets" scenario
   - Documents expected behavior: `ko ls --all` includes both open and closed tickets

3. **testdata/ticket_listing/all_flag.txtar** (new file)
   - Created testscript test validating:
     - Default `ko ls` excludes closed tickets
     - `ko ls --all` includes both open and closed tickets

### Notable Decisions

- **Implementation approach**: The `--all` flag is a boolean that disables the default closed-ticket filter rather than being a status filter value. This is cleaner than treating "all" as a status value since it conceptually represents "no filter" rather than "a specific status".

- **Filter placement**: The check `!*allTickets` was added to both the JSON output path (line 159) and the non-JSON output path (line 179) to ensure consistent behavior across output formats.

- **Test structure**: Created a separate test file (`all_flag.txtar`) rather than extending an existing one, following the pattern of other single-scenario test files in the testdata/ticket_listing/ directory.

### Verification

- `ko ls` (default): Shows only non-closed tickets ✓
- `ko ls --all`: Shows all tickets including closed ones ✓
- `ko ls --help`: Documents the flag ✓
- All testscript tests pass, including the new `all_flag` test ✓
- Full test suite passes (200+ tests) ✓
- File size remains under 500 line limit (286 lines) ✓

### Invariants Compliance

✓ Spec before code: Added spec scenario in ticket_listing.feature
✓ Test for spec: Created all_flag.txtar testscript test
✓ Pure decision logic: Filtering logic is a simple boolean check
✓ File size: cmd_list.go remains at 286 lines (under 500 limit)
✓ Zero external dependencies: No new runtime dependencies added

## Future Considerations

The implementation exactly matches the SSE behavior, which returns all tickets (including closed ones) by default. Users now have three ways to control which tickets appear:
- `ko ls` - Default: excludes closed
- `ko ls --status=closed` - Only closed tickets
- `ko ls --all` - All tickets regardless of status
