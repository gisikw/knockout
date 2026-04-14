## Obligations

1. [observable] `ko agent start` delegates agent spawning to `ko serve` rather than forking a daemon itself.
   Check: Run `ko agent start` from a transient process (e.g. a short-lived shell script or one-shot systemd unit). Confirm the agent process appears as a child of `ko serve` (e.g. via `ps --ppid <ko-serve-pid>` or `systemctl status <ko-serve-unit>`), NOT as a child of PID 1 or the calling process.

2. [observable] The agent process survives the exit of its caller.
   Check: Start an agent via `ko agent start` from a subprocess that immediately exits. Wait a few seconds, then confirm the agent is still running (e.g. `ko agent status` returns running, or the agent log shows continued activity after the caller exited).

3. [observable] `ko serve` exposes a mechanism (IPC, socket, signal, or internal API) that `ko agent start` uses to request agent spawning.
   Check: Trace the communication path — e.g. `strace -e trace=network,ipc ko agent start` or inspect the relevant socket/pipe — and confirm a message is sent to `ko serve` rather than a direct `exec`/`fork` in the `ko agent start` process.

4. [preserved] `ko agent start` still succeeds when `ko serve` is running and an agent is not already active.
   Check: With `ko serve` running, run `ko agent start` and confirm it exits 0 and the agent loop begins (observable via logs or `ko agent status`).

5. [preserved] `ko agent stop` (or equivalent teardown) still terminates the agent cleanly.
   Check: After starting an agent via the new delegation path, run `ko agent stop` and confirm the agent process exits and `ko agent status` reflects a stopped state.
