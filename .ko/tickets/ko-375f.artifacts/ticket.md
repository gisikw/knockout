---
id: ko-375f
status: open
deps: []
created: 2026-02-24T15:44:43Z
type: task
priority: 2
---
# Migrate agent harnesses from YAML templates to shell scripts


## Notes

**2026-02-25 04:03:14 UTC:** Question: What is the primary motivation for migrating from YAML to shell scripts?
Answer: Underlying agent harnesses have different syntax and affordances; using shell scripts is how we can ensure the agent invocation intent carries through, regardless of the actual agent harness

**2026-02-25 04:05:10 UTC:** Question: Which architecture should the implementation use?
Answer: Executable wrapper, though env vars should all be KO_ namespaced

**2026-02-25 04:05:10 UTC:** Question: Should YAML harnesses continue to work after migration?
Answer: Replace: Convert YAML to shell and deprecate YAML
Migrate built-in and user harnesses to shell; YAML support is removed

**2026-02-25 04:05:10 UTC:** Question: Should binary fallback logic stay in Go or move to shell?
Answer: Move to shell: Shell handles fallback logic
Shell scripts use command -v or similar to find the binary
