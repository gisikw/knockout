Looking at this ticket:

**Problem identified:** Agent start/stop messages are currently output to stdout, which clutters stdout and breaks composability.

**Solution defined:** Route these messages to stderr instead, which is standard CLI behavior (status messages on stderr, data on stdout).

**Scope clarified:** Fix both `ko agent start` and `ko agent stop` commands for consistency.

This is a straightforward code change with a clear problem and known fix. The expected output is committed code changes. The ticket doesn't ask "is this a problem?" or "should we investigate?" — it asserts the problem and describes what to build.

```json
{"disposition": "route", "workflow": "task"}
```
