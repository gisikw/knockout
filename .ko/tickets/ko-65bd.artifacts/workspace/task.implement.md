Done. Summary:

**Task 1 — `~/.config/knockout/projects.yml`**: Added `hidden: true` to both the `tickler` and `user` project entries.

Verification passed:
- `ko project ls` omits `tickler` and `user`
- `ko project ls --all` shows them both

No deviations from the plan.
