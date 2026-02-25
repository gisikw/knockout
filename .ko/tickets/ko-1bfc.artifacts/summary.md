# Implementation Summary: Remove 'create' Command Alias

## What Was Done

Successfully removed the 'create' command alias from the knockout CLI, keeping only 'ko add' as the single canonical command for ticket creation.

### Changes Made

1. **main.go (line 26-27)**: Removed the `case "create":` branch from the command dispatcher
2. **cmd_create.go**: Updated all error messages to say "ko add" instead of "ko create"
3. **INVARIANTS.md**: Updated loop invariant documentation to reference only "ko add"
4. **README.md**: Updated loop scope containment section to remove reference to legacy alias
5. **specs/loop.feature**: Removed duplicate test scenario for 'ko create' (kept only 'ko add' scenario)
6. **specs/ticket_creation.feature**: Replaced all 25 instances of "ko create" with "ko add"
7. **specs/ticket_hierarchy.feature**: Replaced all instances of "ko create" with "ko add"
8. **testdata/loop/loop_no_create.txtar**: Updated to test only 'ko add' blocking
9. **testdata/ticket_creation/*.txtar**: Updated all 8 test files to use 'ko add' instead of 'ko create'

## Compliance with Plan

All 7 planned tasks were completed:
- ✅ Removed 'create' case from main.go
- ✅ Help text already correct (only showed 'add', not 'create')
- ✅ Updated specs/ticket_creation.feature
- ✅ Updated testdata/ticket_creation/*.txtar files
- ✅ Updated README.md
- ✅ Updated other testdata files (loop tests)
- ✅ Updated other spec files (loop.feature, ticket_hierarchy.feature)

## Notable Implementation Decisions

1. **Error message consistency**: Updated all error messages in `cmd_create.go` to use "ko add" prefix for consistency with the new single command name. This improves UX by avoiding confusion.

2. **Removed duplicate test scenario**: In specs/loop.feature, there were two nearly identical scenarios testing the blocking behavior - one for 'create' and one for 'add'. Consolidated to a single 'add' scenario since they tested the same underlying function.

3. **Function name unchanged**: The `cmdCreate` function name was intentionally kept as-is. While it contains "Create" in the name, it's an internal implementation detail not exposed to users. Renaming it would require updating tests and wouldn't add user-facing value.

## Verification

- All tests pass: `go test ./...` succeeds
- Build succeeds: `go build` completes without errors
- Command rejected: `ko create` now returns "unknown command 'create'"
- Command works: `ko add` continues to function correctly
- Help output: Only shows 'add' command, no mention of 'create'

## Invariants Check

No invariants were violated. The changes align with:
- **Specs and tests**: All specs were updated alongside their corresponding tests
- **Error handling**: CLI errors properly go to stderr with non-zero exit codes
- **Naming**: Specs describe behavior, not implementation

The implementation is complete, correct, and ready to commit.
