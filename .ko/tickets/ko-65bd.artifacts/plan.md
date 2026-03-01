## Goal
Mark the "user" and "tickler" projects as hidden in the live registry file so they are excluded from `ko project ls` output by default.

## Context
The registry lives at `~/.config/knockout/projects.yml`. It uses the nested format where each project tag has indented properties (`path`, `prefix`, `default`, `hidden`). The `hidden: true` property causes a project to be excluded from `ko project ls` unless `--all` is passed. Both "user" and "tickler" are present in the registry without a `hidden` field. No code changes are needed — this is a data-only edit.

## Approach
Edit `~/.config/knockout/projects.yml` directly, adding `    hidden: true` under the `tickler:` and `user:` entries.

## Tasks
1. [`~/.config/knockout/projects.yml`] — Add `    hidden: true` as a property under both `tickler:` and `user:` project entries.
   Verify: `ko project ls` does not show "tickler" or "user"; `ko project ls --all` shows them both.

## Open Questions
None.
