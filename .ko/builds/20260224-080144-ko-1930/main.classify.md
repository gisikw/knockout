Looking at this ticket, I need to classify the type of work required.

## Analysis

The ticket clearly describes **what to build**:
- Extract hardcoded agent runners from Go code into declarative config files
- Create YAML harness configs in specific locations (both built-in and user-supplied)
- Define the structure for command building, argument templating, environment variable passing
- Update the `agent:` field resolution to use these configs

This is straightforward feature/refactoring work with:
- Clear requirements (what the config files should contain)
- Defined scope (where files should live, what information they describe)
- Specific components to modify (adapter.go, agent field resolution)
- An implementation goal (make agent runners extensible without recompiling)

There's a prerequisite note about ko-f65e, but that's a dependency consideration, not a blocker on classification.

## Classification

This is a **task** — it describes concrete code changes to implement an extensibility feature.

```json
{"disposition": "route", "workflow": "task"}
```
