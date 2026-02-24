## Goal
Enable setting priority via shorthand syntax `ko create -p1 "title"` where the flag accepts 0-4 and sets the ticket priority directly.

## Context
The `ko create` command already supports a `-p` flag for priority (cmd_create.go:27), but it requires the full syntax `-p 1` (with a space). The ticket requests shorthand support for `-p1` (no space), similar to common CLI patterns like `tar -xf` or `grep -i`.

Key findings:
- cmd_create.go:27 defines `priority := fs.Int("p", -1, "priority (0-4)")`
- cmd_create.go:96-98 applies priority if >= 0: `if *priority >= 0 { t.Priority = *priority }`
- cmd_create.go:18-22 uses `reorderArgs()` to normalize flag positions before parsing
- ticket.go:364 shows NewTicket() sets default priority to 2
- Go's flag package already supports both `-p 1` and `-p=1` syntax natively
- The `reorderArgs` function (main.go:88-109) handles flag reordering and knows that "p" is a value-consuming flag

The issue is that `-p1` (without space or equals) is not standard Go flag syntax. Go's flag package expects either `-p 1` or `-p=1`. The request is for `-p1` which looks like a single-character flag followed immediately by its value, which is not parsed correctly by Go's flag package by default.

## Approach
The Go flag package already handles `-p 1` and `-p=1`. Testing shows `-p1` is not supported natively. We need to add preprocessing in `reorderArgs()` to detect `-p<digit>` patterns and transform them to `-p=<digit>` before flag parsing. This keeps the change minimal and localized to the argument preprocessing logic.

## Tasks
1. [main.go:reorderArgs] — Add detection for single-char value flags followed immediately by their value (e.g., `-p1`, `-d"text"`, `-afoo`). When detected, split into `-flag=value` format that Go's flag parser accepts. Insert this logic before the existing flag/positional separation.
   Verify: The transformation correctly handles `-p1` → `-p=1`, `-p0` → `-p=0`, etc.

2. [cmd_create_test.go] — Add test cases covering the shorthand flag syntax: `-p0 "title"`, `-p1 "title"`, `-p4 "title"`. Verify that created tickets have the correct priority field set. Also test that the longform syntax still works (`-p 1`, `-p=1`).
   Verify: `go test -run TestCreate` passes with new test cases.

3. [cmd_create_test.go] — Add a test case for invalid priority values with shorthand syntax (e.g., `-p5`, `-p9`) to ensure they're rejected properly by the existing validation.
   Verify: Test confirms error handling works as expected.

## Open Questions
None — the implementation is straightforward. The existing flag definition, validation, and priority assignment logic all work correctly once we normalize the shorthand syntax to a format Go's flag package accepts.
