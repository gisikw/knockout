Perfect! I've updated the plan to reflect the current state:

## Summary

I've reviewed the existing plan and updated it to reflect the current state of the implementation:

**What's been done:**
- ✅ Task 4: The `.ko/pipeline.yml` has been simplified from the complex shell conditional to just `just test`

**What remains:**
- Tasks 1-3: Core implementation in `build.go` (hasFlakeNix helper, wrapping in runPromptNode and runRunNode)
- Task 5: Remove environment notes from prompts
- Task 6: Add test coverage

**Key updates to the plan:**
1. Marked task 4 as complete with ✅
2. Added context about the error that triggered this ticket (showing the old shell command wasn't being executed properly through `sh -c`)
3. Updated file references to match current line numbers (build.go is 802 lines, not 551)
4. Added more specific implementation guidance for deriving `ticketsDir` in `runRunNode`
5. Referenced the existing `ProjectRoot` function location for implementer convenience

The plan is complete, actionable, and ready for implementation. All open questions have been resolved, and the approach is proven (commit 27d3e1b showed manual wrapping works; we're just automating it).
