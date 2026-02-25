# knockout

A task tracker and build pipeline for autonomous agent workflows. Track tickets,
route work across projects, and burn down backlogs with BFS over the ready queue.

## Usage

```
ko - knockout task tracker

Usage: ko <command> [arguments]

Commands:
  add [title]        Create a new ticket
  show <id>          Show ticket details
  ls                 List open tickets
  ready              Show ready queue (open + deps resolved)
  blocked            Show tickets with unresolved deps
  closed             Show closed tickets

  status <id> <s>    Set ticket status
  start <id>         Set status to in_progress
  close <id>         Set status to closed
  reopen <id>        Set status to open
  block <id>         Set status to blocked
  resolved <id>      Set status to resolved

  dep <id> <dep>     Add dependency
  undep <id> <dep>   Remove dependency
  dep tree <id>      Show dependency tree

  note <id> <text>      Add a note to a ticket
  bump <id>             Touch ticket file to update mtime (reorder within priority)
  query                 Output all tickets as JSONL
  questions <id>        Show plan questions as JSON
  answer <id> <json>    Submit answers to plan questions

  agent build <id>   Run build pipeline against a single ticket
  agent loop         Build all ready tickets until queue is empty
  agent init         Initialize pipeline config in current project
  agent start        Daemonize a loop (background agent)
  agent stop         Stop a running background agent
  agent status       Check if an agent is running

  project set #<tag> [--prefix=p] [--default]
                     Initialize .ko dir, register project, optionally set default
  project ls         List registered projects (default marked with *)

  help               Show this help
  version            Show version
```

### Add options

```
ko add [title] [-d description] [-t type] [-p priority] [-a assignee]
               [--parent id] [--external-ref ref]
               [--design notes] [--acceptance criteria]
               [--tags tag1,tag2]
```

## Concepts

### Ticket IDs

Ticket IDs encode hierarchy. The prefix is derived from existing tickets or,
for new projects, from the directory name (e.g. `my-cool-project` -> `mcp-`):

```
mcp-a1b2            root ticket (depth 0)
mcp-a1b2.c3d4       child (depth 1)
mcp-a1b2.c3d4.e5f6  grandchild (depth 2)
```

Depth is visible by counting dots. Decomposition is bounded by max depth --
at the limit, tickets get blocked for human review instead of further decomposed.

### Statuses

| Status | Meaning |
|--------|---------|
| `captured` | Just captured, needs triage |
| `routed` | Routed to this project from elsewhere |
| `open` | Ready to be worked (eligible for `ko ready`) |
| `in_progress` | Currently being worked |
| `closed` | Done |
| `resolved` | Done, pending human review before close |
| `blocked` | Needs human attention |

`ko ready` only returns `open` and `in_progress` tickets with all deps resolved.
Within the same priority tier, tickets are sorted first by status (`in_progress`
before `open`), then by file modification time (most recently touched first).
Use `ko bump <id>` to move a ticket to the top of its priority tier without
changing its content.

### Build Pipeline

`ko agent build <ticket-id>` runs a workflow-based pipeline against a ticket. The
pipeline config lives in `.ko/config.yaml` (or legacy `.ko/pipeline.yml`) and
declares named **workflows** containing typed **nodes**. Every ticket enters the
`main` workflow.

There are two node types:

- **Decision nodes** return structured JSON dispositions (extracted from the
  last fenced code block in the output). These drive the workflow graph.
- **Action nodes** just do work — their output is not parsed.

Every build outcome removes the ticket from the ready queue:

| Outcome | Effect |
|---------|--------|
| `SUCCEED` | Ticket closed (all nodes in the workflow completed) |
| `FAIL` | Ticket blocked (needs human) |
| `BLOCKED` | Dependency wired, ticket off queue until dep resolves |
| `DECOMPOSE` | Child tickets created, parent blocked on them |

#### Dispositions

Decision nodes end their output with a fenced JSON block. The runner extracts
the **last** fenced JSON block:

```json
{"disposition": "continue"}
```

Valid dispositions:

| Disposition | Meaning | Required fields |
|------------|---------|-----------------|
| `continue` | Advance to next node in workflow | — |
| `fail` | Block ticket for human review | `reason` |
| `blocked` | Wire a dependency | `block_on`, `reason` |
| `route` | Jump to a different workflow | `workflow` |
| `decompose` | Split into subtasks | `subtasks` (array) |
| `resolved` | Mark for human review | `reason` (optional) |

Route targets must be declared in the node's `routes` field — a decision node
cannot route to a workflow it hasn't declared. This prevents prompt injection
from hijacking the workflow graph.

#### Workspace

Each build creates a workspace at `.ko/tickets/<id>.artifacts/workspace/`. Node
outputs are tee'd to `<workflow>.<node>.md` files in the workspace. The artifact
directory persists across builds and after ticket close. Available as
`$KO_TICKET_WORKSPACE` (workspace path) and `$KO_ARTIFACT_DIR` (parent artifact
directory).

### Build Loop

`ko agent loop` burns down the entire ready queue without human intervention. It
builds one ticket at a time, re-querying after each build so that newly
unblocked tickets (from closed deps or decomposition) get picked up.

```bash
ko agent loop                    # run until queue is empty
ko agent loop --max-tickets 5    # stop after 5 tickets
ko agent loop --max-duration 30m # stop after 30 minutes
```

**Scope containment:** During a loop, `ko add` is disabled via the `KO_NO_CREATE`
environment variable. This prevents spawned agents from creating new tickets,
which would cause runaway expansion.

The loop stops when:
- The ready queue is empty (`stopped: empty`)
- `--max-tickets` limit reached (`stopped: max_tickets`)
- `--max-duration` limit reached (`stopped: max_duration`)
- A build execution error occurs (`stopped: build_error`)

Outcome signals (FAIL, BLOCKED, DECOMPOSE) do **not** stop the loop — the
affected ticket is removed from the ready queue and the loop continues.

### Agent Daemon

`ko agent start` daemonizes a loop as a background process, tracking it via
`.ko/agent.pid`. Output is appended to `.ko/agent.log`. Use `ko agent stop` to
terminate and `ko agent status` to check — status includes the last log line
for a quick read on what the agent is doing.

```bash
ko agent start                    # background loop for current project
ko agent start '#myapp'           # background loop for a registered project
ko agent status                   # "running (pid 12345)" + last log line
ko agent stop                     # kill + cleanup
```

Stale PID files (process died) are automatically cleaned up on `start` and `status`.

`ko agent stop` sends SIGTERM (allowing the loop to log the signal and clean
up), then escalates to SIGKILL after 5 seconds if needed. Any in-progress
ticket is reset to `open` and `on_fail` hooks run (worktree cleanup).

### Project Registry

Register projects for cross-project routing:

```bash
ko project set #fort-nix --prefix=nix    # initialize .ko dir, register as "fort-nix"
ko project set #myapp --default           # register and set as default
ko project ls                             # list all registered projects (default marked with *)
```

The `project set` command is an upsert operation: it initializes the `.ko/tickets/` directory if needed, writes the prefix to `.ko/config.yaml`, registers the project in the global registry, and optionally sets it as the default. Running it again on an existing project updates the registration.

Registry lives at `~/.config/knockout/projects.yml`.

## Pipeline Configuration

Pipeline config lives in `.ko/config.yaml`, prompts in `.ko/prompts/`.
Run `ko agent init` to scaffold a starter pipeline, or see `examples/` for
templates.

The config file contains both project settings and pipeline configuration:

```yaml
# .ko/config.yaml
project:
  prefix: ko  # Ticket ID prefix

pipeline:
  model: claude-sonnet-4-5-20250929
  max_retries: 2
  max_depth: 2
  discretion: high

  workflows:
    main:
      - name: triage
        type: decision
        prompt: triage.md
        routes:
          - hotfix
      - name: implement
        type: action
        prompt: implement.md
      - name: verify
        type: action
        run: just test
      - name: review
        type: decision
        prompt: review.md

    hotfix:
      - name: implement
        type: action
        prompt: implement.md

  on_succeed:
    - git add -A
    - git commit -m "ko: implement ${TICKET_ID}"

  on_close:
    - git push
```

**Backwards compatibility:** Legacy `.ko/pipeline.yml` (without `project:` section)
is still supported. The system checks for `config.yaml` first, then falls back to
`pipeline.yml` if not found.

### Pipeline options

| Key | Default | Description |
|-----|---------|-------------|
| `agent` | `claude` | Agent adapter: `claude` \| `cursor` |
| `command` | — | Raw command override (mutually exclusive with `agent`) |
| `allow_all_tool_calls` | `false` | Maps to `--dangerously-skip-permissions` (claude), `--force` (cursor) |
| `allowed_tools` | `[]` | List of tool names to auto-allow (e.g., `Read`, `Write`, `Bash`). Can be set at pipeline, workflow, or node level with override semantics (node > workflow > pipeline). Only used when `allow_all_tool_calls` is false. Tool names are case-sensitive. |
| `model` | — | Default model for all prompt nodes |
| `max_retries` | `2` | Retry attempts per node |
| `max_depth` | `2` | Max decomposition depth |
| `discretion` | `medium` | `low` \| `medium` \| `high` — passed to prompt nodes |
| `step_timeout` | `15m` | Default max duration per pipeline node |

### Node properties

| Key | Required | Description |
|-----|----------|-------------|
| `name` | yes | Node identifier (unique across all workflows) |
| `type` | yes | `decision` or `action` |
| `prompt` | one of | Prompt file in `.ko/prompts/`, or inline text |
| `run` | one of | Shell command to execute |
| `model` | no | Model override for this node |
| `allowed_tools` | no | List of tool names to auto-allow (multiline or inline syntax). Completely replaces any workflow or pipeline level lists. |
| `routes` | no | Workflows this decision node may route to |
| `max_visits` | no | Max times this node can run per build (default: 1) |
| `timeout` | no | Max duration for this node (overrides `step_timeout`) |
| `skills` | no | List of skill directory paths to make available |
| `skill` | no | Skill name — implies prompt "apply /skill-name" |

### Hooks

- **`on_succeed`** runs after all workflows pass, before the ticket is closed.
  Available env: `$TICKET_ID`, `$CHANGED_FILES`, `$KO_TICKET_WORKSPACE`.
- **`on_fail`** runs when a build fails (worktree cleanup, reset state, etc.).
  Best-effort — errors are not propagated. Same env vars available.
- **`on_close`** runs after the ticket is closed. Safe for deploys — if the
  hook kills the process, the ticket is already closed.

### Custom Agent Harnesses

Agent harnesses are executable shell scripts that receive parameters via
KO_-namespaced environment variables and invoke an agent CLI. Built-in harnesses
(`claude`, `cursor`) ship with `ko`. You can add custom harnesses to extend
support for other agents.

Harness search order:
1. `.ko/agent-harnesses/<name>` — project-local overrides (executable file)
2. `~/.config/knockout/agent-harnesses/<name>` — user-global harnesses (executable file)
3. Built-in harnesses (embedded in the `ko` binary)

#### Environment Variables

Shell harnesses receive the following environment variables:

- **`KO_PROMPT`** — The full prompt text to pass to the agent
- **`KO_MODEL`** — The model name (e.g., "sonnet", "opus"), may be empty
- **`KO_SYSTEM_PROMPT`** — System prompt text, may be empty
- **`KO_ALLOW_ALL`** — "true" or "false" indicating whether all tools are allowed
- **`KO_ALLOWED_TOOLS`** — Comma-separated list of allowed tools (e.g., "read,write,bash"), may be empty

#### Example Custom Harness

Example harness (`.ko/agent-harnesses/mycli`):

```bash
#!/bin/sh
# Custom harness for mycli agent
set -e

# Build command arguments
args="--output-format text"

# Add conditional flags only if set
if [ -n "$KO_MODEL" ]; then
  args="$args --model $KO_MODEL"
fi

if [ -n "$KO_SYSTEM_PROMPT" ]; then
  args="$args --system-prompt $KO_SYSTEM_PROMPT"
fi

if [ "$KO_ALLOW_ALL" = "true" ]; then
  args="$args --allow-all"
fi

if [ -n "$KO_ALLOWED_TOOLS" ]; then
  args="$args --allowed-tools $KO_ALLOWED_TOOLS"
fi

# Pass prompt via stdin (or as argument, depending on your agent)
echo "$KO_PROMPT" | mycli $args
```

Make it executable:

```bash
chmod +x .ko/agent-harnesses/mycli
```

Use the custom harness in `.ko/pipeline.yml`:

```yaml
agent: mycli
```

#### Binary Fallback

For agents with multiple binary names, use `command -v` to resolve:

```bash
#!/bin/sh
# Try multiple binary names
if command -v mycli-agent >/dev/null 2>&1; then
  BINARY="mycli-agent"
elif command -v mycli >/dev/null 2>&1; then
  BINARY="mycli"
else
  echo "Error: mycli not found" >&2
  exit 1
fi

# Use $BINARY in your command
echo "$KO_PROMPT" | $BINARY --output-format text
```

### Plan Questions

When a build plan has open questions requiring human input, the pipeline can
block the ticket with structured questions:

```bash
ko block ko-f65e --questions '[{"id":"q1","question":"Tabs or spaces?","options":[{"label":"Spaces","value":"spaces"},{"label":"Tabs","value":"tabs"}]}]'
```

Read questions for a blocked ticket:

```bash
ko questions ko-f65e    # outputs JSON array
```

Answer questions (partial or full):

```bash
ko answer ko-f65e '{"q1":"spaces"}'
```

Each answer is recorded as a note. When all questions are answered, the ticket
automatically unblocks (status returns to `open`).

## Data Model

Tickets are markdown files with YAML frontmatter in `.ko/tickets/`. No database,
no index, no derived state. The file is the source of truth.
