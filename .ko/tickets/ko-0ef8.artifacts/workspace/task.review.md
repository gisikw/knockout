Perfect! The implementation is complete and correct. Let me provide the final disposition:

## Review Complete

The implementation successfully addresses the ticket requirement. The change is minimal, correct, and complete:

- ✓ **Plan followed exactly** - Added git stash command to on_fail hooks as specified
- ✓ **Tests pass** - All Go tests passed during verify stage
- ✓ **No invariants violated** - Configuration change, no code/spec/test invariants apply
- ✓ **Safe and scoped** - Single line addition to pipeline config, no security concerns
- ✓ **Complete solution** - Stash preserves failed work while cleaning worktree for next build

The implementation adds `git stash push -u -m "ko: ${TICKET_ID} failed build"` to the on_fail hook list in `.ko/pipeline.yml`. This ensures that when builds fail, uncommitted changes (including untracked files) are stashed with a descriptive message containing the ticket ID, leaving a clean worktree for subsequent builds while preserving the failed work for potential manual inspection or retry.

```json
{"disposition": "continue"}
```
