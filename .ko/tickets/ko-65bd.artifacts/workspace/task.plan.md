**Plan summary:**

The ticket asks to mark the `user` and `tickler` projects as hidden in the live registry. This is a pure data edit — no code changes needed.

**What I found:**
- The registry is at `~/.config/knockout/projects.yml`
- Both `tickler` and `user` are present without a `hidden` field
- Adding `    hidden: true` under each entry is all that's required
- `ko project ls` already filters hidden projects; `ko project ls --all` reveals them

**Task:**
1. Add `    hidden: true` to the `tickler:` and `user:` entries in `~/.config/knockout/projects.yml`
   - Verify: `ko project ls` omits them; `ko project ls --all` shows them
