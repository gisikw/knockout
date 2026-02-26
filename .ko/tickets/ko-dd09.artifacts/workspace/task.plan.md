Plan written! The fix is clear and well-scoped:

**Summary**: The issue is that decision nodes receive their own previous classification output as "prior context" on retry builds, confusing Haiku. The fix is a simple conditional in `build.go` to only inject prior context for action nodes (which benefit from continuity) and not decision nodes (which should make fresh, idempotent routing decisions). Three tasks: add the conditional logic, write a test to verify the behavior, and confirm existing integration tests pass.
