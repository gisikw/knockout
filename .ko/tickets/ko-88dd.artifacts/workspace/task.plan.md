Great! The plan has been updated. The open question has been resolved - the ticket author confirmed that `require_clean_tree: true` should use the standard interpretation: **prevent builds when there ARE uncommitted changes** (dirty tree), not when the tree is clean.

The plan is now complete and ready for implementation:

## Summary

The plan adds `require_clean_tree: true` support to pipeline.yml with the following approach:

**Confirmed Behavior**: When `require_clean_tree: true` is set, builds will be blocked if there are uncommitted changes in the working tree (outside `.ko/`). This is the standard Git interpretation of "require clean tree."

**Implementation Strategy**:
1. Add `RequireCleanTree` field to Pipeline struct and parse it from YAML
2. Create `isWorkingTreeClean()` helper that shells out to `git status --porcelain` and filters out `.ko/` changes
3. Enhance `BuildEligibility()` to check working tree status when the flag is enabled
4. Add comprehensive tests covering dirty tree rejection and `.ko/` directory exclusion
5. Document the option in example pipeline files

The plan is well-structured with 8 specific tasks, each with verification steps. All open questions have been resolved.
