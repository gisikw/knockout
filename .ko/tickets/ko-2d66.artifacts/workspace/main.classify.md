knockout — just --list for recipes
```json
{"disposition": "route", "workflow": "task"}
```

**Reasoning:**

The ticket describes a well-diagnosed issue with a clear architectural fix:
- **Root cause identified**: systemd cgroup reaping kills agents when ko agent start runs in one-shot services
- **Solution specified**: implement delegation to ko serve instead of double-fork daemonization
- **Implementation path clear**: ko agent start should signal ko serve to spawn agents, keeping them in ko serve's long-lived cgroup

This is a code change task with known requirements, not a bug requiring diagnosis or research with uncertain solutions.
