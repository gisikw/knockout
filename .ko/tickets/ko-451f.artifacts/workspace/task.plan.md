The plan is ready for review at `.ko/tickets/ko-451f.artifacts/plan.md`. 

**Summary:** This ticket updates the README to remove documentation for the removed `ko triage` and `ko blocked` commands, replacing them with the consolidated `ko update` command. The changes include:

1. Removing the `ko blocked` and `ko triage` command lines from the Commands section
2. Removing their JSON output documentation
3. Updating the "Plan Questions" examples to use `ko update` instead of `ko triage`
4. Adding proper documentation for the `ko update` command
5. Running `just install` to rebuild and install the binary

The plan is straightforward with no open questions—all dependent tickets are complete and the changes are simple find-and-replace operations with the addition of proper `ko update` documentation.
