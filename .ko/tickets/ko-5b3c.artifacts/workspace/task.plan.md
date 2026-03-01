Plan updated. Here's a summary of what changed from the prior plan:

**What was answered:** The open question about `--all` is resolved — hidden projects are revealed via `ko project ls --all`.

**Key plan changes:**
- Task 1: Added a third Gherkin scenario for `ko project ls --all` showing hidden projects.
- Task 6 (`cmdProjectLs`): Now includes adding `--all` flag to `flag.NewFlagSet`, filtering logic keyed on `*allProjects`, adding `IsHidden bool` to `projectJSON` for `--all` JSON output, and updating the usage string.
- Task 8: Added `TestCmdProjectLsAllShowsHidden` unit test.
- Task 9 (testscript): Now also exercises `ko project ls --all` and asserts the hidden project appears.
- Open Questions: None remaining.

The build failure ("unknown workflow 'main'") was a transient system error from the prior run — it doesn't affect the plan.
