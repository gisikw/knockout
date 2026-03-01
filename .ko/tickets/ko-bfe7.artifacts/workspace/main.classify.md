Looking at this ticket:

**What the ticket asks for:**
- Restructure projects.yml to reduce redundancy
- Aggregate path, default, and prefix under project names
- Update all code that depends on this file structure

**Key characteristics:**
- Describes **what to build** (a refactored configuration structure)
- Has a clear target state (consolidated fields under project names)
- Specifies the scope of changes (everywhere with dependencies on this file)
- Expected output is committed code

This is not asking "can we do this?" (research), and it's not reporting something broken (bug). It's a clear refactoring task with a defined approach.

```json
{"disposition": "route", "workflow": "task"}
```
