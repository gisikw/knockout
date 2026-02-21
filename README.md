# knockout

A task tracker and build pipeline for autonomous agent workflows. Track tickets,
route work across projects, and burn down backlogs with BFS over the ready queue.

## Usage

```
ko - knockout task tracker

Usage: ko <command> [arguments]

Commands:
  create [title]     Create a new ticket
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

  dep <id> <dep>     Add dependency
  undep <id> <dep>   Remove dependency
  dep tree <id>      Show dependency tree

  link <id1> <id2>   Link two tickets
  unlink <id1> <id2> Unlink two tickets

  add-note <id> <text>  Add a note to a ticket
  bump <id>             Touch ticket file to update mtime (reorder within priority)
  query                 Output all tickets as JSONL

  init <prefix>      Initialize project with ticket prefix

  agent build <id>   Run build pipeline against a single ticket
  agent loop         Build all ready tickets until queue is empty
  agent init         Initialize pipeline config in current project
  agent start        Daemonize a loop (background agent)
  agent stop         Stop a running background agent
  agent status       Check if an agent is running

  add '<title> [#tag]'  Capture a task, route by tag if registered
  register #<tag>    Register current project in the global registry
  default [#<tag>]   Show or set the default project for routing
  projects           List registered projects

  help               Show this help
  version            Show version
```

### Create options

```
ko create [title] [-d description] [-t type] [-p priority] [-a assignee]
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
| `blocked` | Needs human attention |

`ko ready` only returns `open` and `in_progress` tickets with all deps resolved.
Within the same priority tier, tickets are sorted by file modification time
(most recently touched first). Use `ko bump <id>` to move a ticket to the top
of its priority tier without changing its content.

### Build Pipeline

`ko agent build <ticket-id>` runs a workflow-based pipeline against a ticket. The
pipeline config lives in `.ko/pipeline.yml` and declares named **workflows**
containing typed **nodes**. Every ticket enters the `main` workflow.

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

Route targets must be declared in the node's `routes` field — a decision node
cannot route to a workflow it hasn't declared. This prevents prompt injection
from hijacking the workflow graph.

#### Workspace

Each build creates a workspace at `.ko/builds/<ts>-<id>/workspace/`. Node
outputs are tee'd to `<workflow>.<node>.md` files in the workspace. The path
is available to all nodes and hooks as `$KO_TICKET_WORKSPACE`.

### Build Loop

`ko agent loop` burns down the entire ready queue without human intervention. It
builds one ticket at a time, re-querying after each build so that newly
unblocked tickets (from closed deps or decomposition) get picked up.

```bash
ko agent loop                    # run until queue is empty
ko agent loop --max-tickets 5    # stop after 5 tickets
ko agent loop --max-duration 30m # stop after 30 minutes
```

**Scope containment:** During a loop, `ko create` and `ko add` are disabled
via the `KO_NO_CREATE` environment variable. This prevents spawned agents from
creating new tickets, which would cause runaway expansion.

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
ko agent stop                     # SIGTERM + cleanup
```

Stale PID files (process died) are automatically cleaned up on `start` and `status`.

### Project Registry

Register projects for cross-project routing:

```bash
ko register #fort-nix     # register current project as "fort-nix"
ko register #myapp         # register another project
ko default #myapp          # set default for unrecognized tags
ko projects                # list all registered projects
```

Registry lives at `~/.config/knockout/projects.yml`.

## Pipeline Configuration

Pipeline config lives in `.ko/pipeline.yml`, prompts in `.ko/prompts/`.
Run `ko agent init` to scaffold a starter pipeline, or see `examples/` for
templates.

```yaml
# .ko/pipeline.yml
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

### Pipeline options

| Key | Default | Description |
|-----|---------|-------------|
| `agent` | `claude` | Agent adapter: `claude` \| `cursor` |
| `command` | — | Raw command override (mutually exclusive with `agent`) |
| `allow_all_tool_calls` | `false` | Maps to `--dangerously-skip-permissions` (claude), `--force` (cursor) |
| `model` | — | Default model for all prompt nodes |
| `max_retries` | `2` | Retry attempts per node |
| `max_depth` | `2` | Max decomposition depth |
| `discretion` | `medium` | `low` \| `medium` \| `high` — passed to prompt nodes |

### Node properties

| Key | Required | Description |
|-----|----------|-------------|
| `name` | yes | Node identifier (unique across all workflows) |
| `type` | yes | `decision` or `action` |
| `prompt` | one of | Prompt file in `.ko/prompts/` |
| `run` | one of | Shell command to execute |
| `model` | no | Model override for this node |
| `routes` | no | Workflows this decision node may route to |
| `max_visits` | no | Max times this node can run per build (default: 1) |

### Hooks

- **`on_succeed`** runs after all workflows pass, before the ticket is closed.
  Available env: `$TICKET_ID`, `$CHANGED_FILES`, `$KO_TICKET_WORKSPACE`.
- **`on_close`** runs after the ticket is closed. Safe for deploys — if the
  hook kills the process, the ticket is already closed.

## Data Model

Tickets are markdown files with YAML frontmatter in `.ko/tickets/`. No database,
no index, no derived state. The file is the source of truth.
