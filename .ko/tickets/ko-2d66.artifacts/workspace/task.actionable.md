knockout — just --list for recipes
I'll evaluate this plan against the three criteria.

## Assessment

**Open Questions**: The "Open Questions" section lists three items, but Q1 and Q2 are documented decision notes ("No action needed", "Confirmed no action needed"). However, **Q3 is a genuine open question** that needs human input:
- The plan asks whether to write just the port number or the full address to `serve.addr`
- It recommends the full address form, but this is a design decision worth confirming before implementation
- The choice affects both Task 1 (what to write) and Task 2 (what format the delegation function expects to parse)

**Completeness**: ✓ The plan is well-structured with clear goal, comprehensive context, specific approach, and 5 well-defined tasks with verification steps. Task ordering is logical (server lifecycle → delegation function → agent integration → testing → spec).

**Feasibility**: ✓ Referenced files (`cmd_agent.go`, `cmd_serve.go`, `mutation.go`) are documented as existing. The HTTP delegation pattern is noted as already available in the `/ko` endpoint. The approaches (XDG_STATE_HOME path variable, httptest mocking, feature specs) are all standard and plausible.

```json
{"disposition": "needs_input", "plan_questions": [{"id": "q1", "question": "What format should ko serve write to the serve.addr file?", "context": "Task 1 writes the listening address; Task 2's HTTP client must read and use this value. The plan recommends the full address form (e.g., '127.0.0.1:19876') for convenience.", "options": [{"label": "Full address (Recommended)", "value": "full_address", "description": "Write host and port together via net.JoinHostPort(), ready to use directly in HTTP client URLs"}, {"label": "Port only", "value": "port_only", "description": "Write just the port number, client reconstructs localhost:port"}]}]}
```
