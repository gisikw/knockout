---
id: ko-2d66
status: open
deps: []
created: 2026-03-18T17:40:45Z
type: task
priority: 2
---
# ko agent: delegate to ko serve instead of double-fork

Double-fork daemonization doesn't survive systemd cgroup reaping. When ko agent start runs inside a one-shot service (e.g. morning briefing), the agent process gets SIGTERM'd when the one-shot exits because systemd kills the entire cgroup regardless of reparenting.

Fix: ko agent start should signal ko serve (which runs in its own long-lived service) to spawn the agent loop. The agent then lives in ko serve's cgroup and survives any caller's lifecycle.

Discovery: res-c39d got killed 23 seconds into a build because the briefing one-shot exited and reaped the cgroup.
