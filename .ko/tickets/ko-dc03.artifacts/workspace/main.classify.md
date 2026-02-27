Looking at this ticket:

**Title**: Agent loop should detect flake.nix and wrap harness invocations in 'nix develop --command' automatically, so pipelines don't need to handle nix environments themselves

**Analysis**:
- The ticket describes a **specific feature to build**: automatically detect flake.nix and wrap harness invocations
- It specifies the **desired behavior**: pipelines shouldn't need to handle nix environments themselves
- It has a **clear implementation direction**: detect flake.nix and use `nix develop --command`
- The expected output is **code changes** that add this automatic wrapping capability

This is a feature request with a defined scope and solution approach, not a question requiring investigation or a report of broken behavior.

```json
{"disposition": "route", "workflow": "task"}
```
