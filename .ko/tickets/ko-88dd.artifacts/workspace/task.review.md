The implementation is complete and correct. All 11 planned tasks were completed:

- `RequireCleanTree` field added to `Pipeline` struct and parsed from YAML
- `isWorkingTreeClean` helper shells out to `git status --porcelain`, filtering `.ko/` paths
- `BuildEligibility` signature updated and clean-tree gate added
- `RunBuild` and `cmd_build.go` call sites updated (with a sensible reorder in `cmd_build.go` to load pipeline before checking eligibility)
- Unit tests in `build_test.go` and `pipeline_test.go` all pass
- Gherkin spec scenarios added to `specs/pipeline.feature`
- `.txtar` integration test added and passing
- Commented examples added to all three example pipelines

No invariant violations: the git dependency exception is explicitly confirmed by the ticket author and documented in the plan. All tests pass.

```json
{"disposition": "continue"}
```
