---
id: ko-dc03
status: open
deps: []
created: 2026-02-27T15:15:31Z
type: task
priority: 2
---
# Agent loop should detect flake.nix and wrap harness invocations in 'nix develop --command' automatically, so pipelines don't need to handle nix environments themselves

## Notes

**2026-02-27 15:30:17 UTC:** ko: FAIL at node 'verify' — node 'verify' failed after 3 attempts: command failed: exit status 127
sh: line 1: if [ -f flake.nix ]; then nix develop --command just test; else just test; fi: command not found

