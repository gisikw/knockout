knockout — just --list for recipes
**Obligations summary:**

1. **[observable]** `ko agent start` delegates to `ko serve` for spawning — agent appears as child of `ko serve`, not the caller.
2. **[observable]** Agent survives caller exit — still running after the invoking process terminates.
3. **[observable]** An IPC/socket/API mechanism exists between `ko agent start` and `ko serve` — no direct fork in the start command.
4. **[preserved]** `ko agent start` still succeeds normally when `ko serve` is running.
5. **[preserved]** `ko agent stop` still cleanly terminates the agent regardless of the new spawn path.
