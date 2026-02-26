Perfect! I've written a comprehensive plan for supporting `#tag` as shorthand for `--project=tag`. The plan:

1. **Goal**: Enable `#tag` syntax as a convenient shorthand for `--project=tag` across all commands that support project filtering.

2. **Context**: I found that:
   - The infrastructure already exists (`CleanTag()` function)
   - `cmd_project.go` already uses this pattern
   - All cross-project commands funnel through `resolveProjectTicketsDir()`
   - This means we only need to modify one function to enable the feature everywhere

3. **Approach**: Extend the existing `resolveProjectTicketsDir()` function to recognize positional arguments starting with `#` as project tags.

4. **Tasks**:
   - Modify `resolveProjectTicketsDir()` to detect and handle `#tag` args
   - Add comprehensive test coverage for the new syntax
   - Add a behavioral spec documenting the feature
   - Verify existing project set tests already work with `#tag`

5. **Open Questions**: None - the implementation is straightforward and follows existing patterns in the codebase.
