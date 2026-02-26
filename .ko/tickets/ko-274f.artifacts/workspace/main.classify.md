Looking at this ticket, I need to classify the work it describes.

The ticket is discussing a potential need for a way to programmatically indicate if the working directory is clean, to know whether it's safe to spin up an agent. The author is wondering about possible solutions: exposing something via `ko agent status`, or adding a pre-loop git stash step to `pipeline.yml`.

**Analysis:**
- The ticket is phrased as a question/exploration: "I feel like we may need...", "Wonder if...", "Or perhaps..."
- It's asking whether something is necessary and exploring what the solution might look like
- The expected output is investigation and findings, not committed code
- This is discovery work to determine the right approach before implementing anything

This is **research** - the author is investigating whether this capability is needed and what form it should take, rather than describing what code needs to be built.

```json
{"disposition": "route", "workflow": "research"}
```
