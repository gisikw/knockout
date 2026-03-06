---
id: ko-65bd
status: closed
deps: [ko-5b3c]
created: 2026-03-01T06:19:34Z
type: task
priority: 2
---
# Please update the existing registry to make the "user" and "tickler" projects hidden

## Notes

**2026-03-01 13:56:06 UTC:** ## Summary

Updated `~/.config/knockout/projects.yml` to mark the "user" and "tickler" projects as hidden by adding `hidden: true` under each project's entry.

### What was done
- Added `hidden: true` to the `tickler:` entry in the registry
- Added `hidden: true` to the `user:` entry in the registry

### Notable decisions
- This was a data-only change with no code modifications required, as expected by the plan.
- The registry file lives outside the git repository (`~/.config/knockout/projects.yml`), so the change does not appear in the git diff — this is expected behavior.

### Verification
- `ko project ls` correctly omits both "tickler" and "user" from output
- `ko project ls --all` correctly shows both hidden projects alongside visible ones

**2026-03-01 13:56:06 UTC:** ko: SUCCEED
