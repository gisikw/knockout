The plan was already complete. The only open question — whether to transfer or clear the default when evicting a stale tag — has been answered in the ticket notes: **transfer to the new tag**. I've updated the plan to mark it resolved.

The plan is ready to implement:
1. **`cmd_project.go:110`** — Add a pre-eviction loop before `reg.Projects[tag] = root` that removes any existing entry pointing to the same `root` under a different tag, cleans up `reg.Prefixes`, and transfers `reg.Default` if it pointed to the evicted tag.
2. **`cmd_project_test.go`** — Add `TestCmdProjectSetRetagEvictsOldTag` covering the retag scenario and the default-transfer sub-case.
