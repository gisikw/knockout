# Findings: Programmatically Checking if Working Directory is Clean

## Executive Summary

The codebase currently has no mechanism to check if the working directory is clean before starting an agent or loop. However, there are two viable approaches to address this:

1. **Extend `ko agent status` to expose git cleanliness** (recommended for programmatic checks)
2. **Add a pre-loop hook system to `pipeline.yml`** (recommended for automated cleanup)

Both approaches are complementary and could be implemented together.

## Current State

### Agent Status Command

**File:** `cmd_agent.go:259-343`

The `ko agent status` command currently exposes:
- `provisioned`: whether `.ko/config.yaml` or `.ko/pipeline.yml` exists
- `running`: whether an agent is running
- `pid`: process ID of running agent
- `last_log`: last line from `.ko/agent.log`

**Output formats:**
- Human-readable: `running (pid 12345)` with last log line
- JSON: structured output with `--json` flag

### Loop Execution Flow

**File:** `cmd_loop.go:83-221` and `loop.go:66-139`

The agent loop:
1. Acquires exclusive lock via `acquireAgentLock()` to prevent concurrent runs
2. Sets `KO_NO_CREATE=1` environment variable
3. Iterates through ready tickets via `ReadyQueue()`
4. For each ticket:
   - Marks as `in_progress` (build.go:89)
   - Executes workflow
   - Runs `on_succeed` or `on_fail` hooks

**No working directory checks are performed before loop starts.**

### Existing Git Integration

**Files:** `.ko/pipeline.yml:66-71`, `build.go:106-120`, `build.go:149-159`

The system uses git in several places:

1. **`on_succeed` hooks** (runs after all workflow nodes succeed, before closing ticket):
   ```yaml
   on_succeed:
     - git add -A && git commit -m "ko: ${TICKET_ID}"
   ```

2. **`on_fail` hooks** (runs when build fails):
   ```yaml
   on_fail:
     - git add .ko/tickets/ && git diff --cached --quiet || git commit -m "ko: ${TICKET_ID} (blocked)"
     - git stash push -u -m "ko: ${TICKET_ID} failed build"
   ```

The second `on_fail` hook was added by ticket ko-0ef8 to prevent dirty worktree issues between builds.

### Hook System Architecture

**File:** `pipeline.go:35-37`

The pipeline config supports three hook types:
- `on_succeed`: runs after all stages pass (before close)
- `on_fail`: runs on build failure
- `on_close`: runs after ticket is closed

**No pre-build or pre-loop hooks exist.**

## Problem Context

From ticket ko-0ef8 (closed), there was a similar issue:
> "When a build fails at verify (or any node), the agent leaves uncommitted code changes in the working tree. This prevents clean builds of subsequent tickets and creates confusion about provenance."

The solution was to add `git stash` to `on_fail` hooks. However, this addresses cleanup *after* failure, not prevention *before* starting.

The current ticket (ko-274f) asks:
> "I feel like we may need a way to indicate if the working dir is clean, programmatically. From the perspective of knowing whether it's safe to spin up an agent."

This suggests two use cases:
1. **External orchestration** - scripts/tools checking if it's safe to start an agent
2. **Pre-loop safety** - agent refusing to start if working directory is dirty

## Recommended Solutions

### Option 1: Extend `ko agent status` (Recommended for External Checks)

**Implementation:**

Add a `working_dir_clean` field to the status JSON output that checks for uncommitted changes.

**Changes required:**

1. **cmd_agent.go:259-264** - Add field to `agentStatusJSON` struct:
   ```go
   type agentStatusJSON struct {
       Provisioned     bool   `json:"provisioned"`
       Running         bool   `json:"running"`
       Pid             int    `json:"pid,omitempty"`
       LastLog         string `json:"last_log,omitempty"`
       WorkingDirClean bool   `json:"working_dir_clean"`  // NEW
   }
   ```

2. **cmd_agent.go:266-343** - Add git status check in `cmdAgentStatus()`:
   ```go
   // After checking provisioned status (line ~295), add:
   status.WorkingDirClean = isWorkingDirClean(ticketsDir)
   ```

3. **Add helper function:**
   ```go
   // isWorkingDirClean checks if the git working directory has uncommitted changes.
   func isWorkingDirClean(ticketsDir string) bool {
       projectRoot := ProjectRoot(ticketsDir)
       cmd := exec.Command("git", "status", "--porcelain")
       cmd.Dir = projectRoot
       output, err := cmd.Output()
       if err != nil {
           return false // Treat git errors as "not clean"
       }
       // Empty output means clean working directory
       return len(bytes.TrimSpace(output)) == 0
   }
   ```

**Human-readable output:**
```
running (pid 12345)
  last: loop: building ko-274f
  working directory: clean
```

**JSON output:**
```json
{
  "provisioned": true,
  "running": true,
  "pid": 12345,
  "last_log": "loop: building ko-274f",
  "working_dir_clean": true
}
```

**Benefits:**
- Minimal changes to existing code
- Backwards compatible (JSON consumers can ignore new field)
- Enables external orchestration scripts: `if $(ko agent status --json | jq -r .working_dir_clean); then ko agent start; fi`
- No behavioral changes to agent itself

**Considerations:**
- Does not enforce cleanliness, only reports it
- External scripts must check before starting agent
- Git must be available and working directory must be a git repository

---

### Option 2: Add Pre-Loop Hook to `pipeline.yml` (Recommended for Automated Cleanup)

**Implementation:**

Add a new `on_start` or `before_loop` hook that runs once when `ko agent loop` or `ko agent start` begins.

**Changes required:**

1. **pipeline.go:35-37** - Add new hook field:
   ```go
   type Pipeline struct {
       // ... existing fields ...
       OnStart     []string  // shell commands to run before loop starts
       OnSucceed   []string
       OnFail      []string
       OnClose     []string
   }
   ```

2. **pipeline.go:176-189** - Parse new section in YAML parser:
   ```go
   if trimmed == "on_start:" {
       section = "on_start"
       continue
   }
   ```

3. **pipeline.go:347-361** - Add parsing for on_start commands:
   ```go
   case "on_start":
       if strings.HasPrefix(trimmed, "- ") {
           cmd := strings.TrimPrefix(trimmed, "- ")
           p.OnStart = append(p.OnStart, cmd)
       }
   ```

4. **cmd_loop.go:83-221** - Run hooks before loop starts:
   ```go
   func cmdAgentLoop(args []string) int {
       // ... existing setup code ...

       // Run on_start hooks before loop begins (after line ~130)
       projectRoot := ProjectRoot(ticketsDir)
       if len(p.OnStart) > 0 {
           if err := runStartHooks(ticketsDir, p.OnStart, projectRoot); err != nil {
               fmt.Fprintf(os.Stderr, "ko agent loop: on_start hook failed: %v\n", err)
               return 1
           }
       }

       // ... rest of loop execution ...
   }
   ```

5. **Add helper function in build.go or hooks.go:**
   ```go
   // runStartHooks executes on_start commands before the loop begins.
   // Unlike on_fail/on_succeed, these run without ticket context.
   func runStartHooks(ticketsDir string, cmds []string, projectRoot string) error {
       for _, cmd := range cmds {
           if cmd == "" {
               continue
           }
           // No ticket context for on_start hooks
           expanded := cmd

           execCmd := exec.Command("sh", "-c", expanded)
           execCmd.Dir = projectRoot
           execCmd.Stdout = os.Stdout
           execCmd.Stderr = os.Stderr

           if err := execCmd.Run(); err != nil {
               return fmt.Errorf("hook failed: %s: %v", cmd, err)
           }
       }
       return nil
   }
   ```

**Example pipeline.yml:**

```yaml
model: claude-sonnet-4-5-20250929
max_retries: 2
max_depth: 2
discretion: medium

workflows:
  # ... workflows ...

# Run before loop starts (fail-fast if working dir is dirty)
on_start:
  - git diff --quiet && git diff --cached --quiet || (echo "Working directory is not clean, refusing to start loop" && exit 1)
  # Or auto-stash: git stash push -u -m "ko: auto-stash before loop"

on_succeed:
  - git add -A && git commit -m "ko: ${TICKET_ID}"

on_fail:
  - git add .ko/tickets/ && git diff --cached --quiet || git commit -m "ko: ${TICKET_ID} (blocked)"
  - git stash push -u -m "ko: ${TICKET_ID} failed build"
```

**Benefits:**
- Provides declarative safety checks in pipeline config
- Users can choose behavior: fail-fast, auto-stash, or warn
- Consistent with existing hook architecture
- Works for both `ko agent loop` and `ko agent start`

**Considerations:**
- Requires pipeline config changes by users
- Adds complexity to loop startup
- Hook failures will prevent agent from starting
- No ticket context available (hooks run before any ticket is selected)

---

### Option 3: Hybrid Approach (Best Overall Solution)

Implement **both** Option 1 and Option 2:

1. **`ko agent status --json`** exposes `working_dir_clean` for programmatic queries
2. **`on_start` hooks** allow users to enforce policies declaratively

This provides:
- **External orchestration:** Scripts can query status before starting agent
- **Internal enforcement:** Pipeline can enforce cleanliness policy per-project
- **Flexibility:** Users choose their workflow (fail-fast vs auto-stash vs permissive)

**Example external script:**
```bash
#!/bin/bash
# scripts/safe-start-agent.sh
STATUS=$(ko agent status --json)

if ! echo "$STATUS" | jq -e '.working_dir_clean' > /dev/null; then
  echo "Working directory is dirty. Stashing changes..."
  git stash push -u -m "auto-stash before agent start"
fi

ko agent start
```

**Example pipeline config:**
```yaml
on_start:
  # Fail fast if working dir is dirty (strict mode)
  - git diff --quiet && git diff --cached --quiet || (echo "ERROR: Working directory is dirty" && exit 1)
```

---

## Related Tickets and Context

- **ko-0ef8** (closed): "Agent should not leave dirty worktree on build failure"
  - Solution: Added `git stash` to `on_fail` hooks
  - This prevents dirty state *between* builds in a loop
  - Does not prevent dirty state *before* loop starts

- **ko-fd4e** (closed): Added `.ko/.gitignore` to exclude runtime files
  - Prevents `agent.lock`, `agent.pid`, `agent.log` from showing as dirty

- **ko-1390** (research): "Should agent.lock and agent.pid be .gitignored?"
  - Led to ko-fd4e implementation

## Code Locations Reference

- `cmd_agent.go:259-343` - `cmdAgentStatus()` and `agentStatusJSON` struct
- `cmd_loop.go:83-221` - `cmdAgentLoop()` main loop entry point
- `loop.go:66-139` - `RunLoop()` core loop logic
- `build.go:57-220` - `RunBuild()` individual ticket build
- `pipeline.go:23-40` - `Pipeline` struct with hook fields
- `pipeline.go:130-392` - YAML parsing logic
- `.ko/pipeline.yml:66-71` - Current hook configuration

## Testing Recommendations

If implementing Option 1 (status extension):
1. Test with clean working directory
2. Test with uncommitted changes
3. Test with staged changes
4. Test in non-git directory (should report false)
5. Test JSON output format
6. Test backwards compatibility of human-readable output

If implementing Option 2 (on_start hooks):
1. Test hook success (clean working dir)
2. Test hook failure (dirty working dir with fail-fast hook)
3. Test auto-stash behavior
4. Test multiple hooks in sequence
5. Test empty/missing on_start section
6. Test hook failure prevents loop from starting
7. Integration test: `ko agent start` with on_start hooks

## Recommended Actions

### Immediate (High Priority)

1. **Implement Option 1**: Extend `ko agent status` with `working_dir_clean` field
   - Low risk, high value for external orchestration
   - Enables users to build their own safety checks immediately
   - Estimated effort: ~1 hour

### Short Term (Medium Priority)

2. **Implement Option 2**: Add `on_start` hooks to pipeline.yml
   - Provides declarative safety enforcement
   - Consistent with existing hook architecture
   - Estimated effort: ~2-3 hours

3. **Update documentation**: Add examples of using both features
   - Document `ko agent status --json` output format
   - Provide example `on_start` hooks for common patterns
   - Add section on working directory hygiene best practices

### Long Term (Nice to Have)

4. **Consider additional pre-loop checks**:
   - Check for stale `.ko/builds/` directories
   - Verify git branch is not detached HEAD
   - Check for uncommitted changes in `.ko/tickets/` only (allow other changes)
   - Warn if on unstable branch (e.g., not main/master)

5. **Add `--force` flag to `ko agent loop`** to bypass on_start hooks for emergency repairs

---

## Conclusion

The codebase does not currently provide a way to check if the working directory is clean before starting an agent. Both suggested solutions (extending `ko agent status` and adding pre-loop hooks) are viable and complementary.

**Recommended approach:** Implement both as a hybrid solution. This provides maximum flexibility for users while maintaining backwards compatibility.

The `ko agent status` extension is simpler and lower risk, making it a good first step. The `on_start` hooks require more integration work but provide better ergonomics for teams that want automatic enforcement.
