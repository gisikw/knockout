Perfect! I've created the implementation plan. The plan identifies:

**What needs to be done:**
- Add the `PlanQuestions` field to the `showJSON` struct in cmd_show.go
- Populate that field when constructing the JSON output
- Test the change manually

**Key insights from the investigation:**
- The `ko ls --json` command already includes plan questions via the `ticketJSON` struct
- The `ko show --json` command uses a different struct (`showJSON`) that's missing this field
- The `PlanQuestion` type is already defined and properly integrated into the system
- This is a simple consistency fix to align the two JSON outputs

The plan is concise and proportional to the task, which is a straightforward single-field addition with no architectural complexity.
