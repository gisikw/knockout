knockout — just --list for recipes
Looking at the ticket:

**"Add the ability to specify pipeline parallelism - some tickets (like research) can be run independently rather than sequentially, and we don't have to enforce the same restrictions as with other work."**

This describes **what to build**: a feature that enables specifying pipeline parallelism so certain tickets can run independently. The expected output is code implementing this capability.

```json
{"disposition": "route", "workflow": "task"}
```
