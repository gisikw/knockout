Perfect! I've written a comprehensive plan to the artifact directory. The plan includes:

## Summary

I've created a detailed implementation plan for adding `require_clean_tree: true` support to pipeline.yml. The plan covers:

**Goal**: Add a pipeline configuration option to prevent builds from starting when the working tree has uncommitted changes (outside `.ko/`)

**Key Changes**:
1. Add `RequireCleanTree` field to the Pipeline struct
2. Parse the field from YAML configuration
3. Implement `isWorkingTreeClean()` helper that checks git status while ignoring `.ko/` directory
4. Enhance `BuildEligibility()` to check working tree status when the flag is enabled
5. Add comprehensive tests for the new functionality
6. Document the option in example pipeline files

**Open Question**: I flagged a potential inconsistency in the ticket wording. The ticket says "prevent if clean" but the config name `require_clean_tree: true` and standard Git patterns suggest it should "prevent if dirty." I've planned the implementation using the standard interpretation (prevent builds when there are uncommitted changes), but noted this should be clarified if the opposite behavior was intended.

The plan is now ready at `.ko/tickets/ko-88dd.artifacts/plan.md` for review.
