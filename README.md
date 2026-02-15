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
  query                 Output all tickets as JSONL

  build <id>         Run build pipeline against ticket

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

### Build Pipeline

`ko build <ticket-id>` runs a YAML-defined pipeline against a ticket. Every
outcome removes the ticket from the ready queue:

| Outcome | Effect |
|---------|--------|
| `SUCCEED` | Ticket closed |
| `FAIL` | Ticket blocked (needs human) |
| `BLOCKED` | Dependency wired, ticket off queue until dep resolves |
| `DECOMPOSE` | Child tickets created, parent blocked on them |

### Project Registry

Register projects for cross-project routing:

```bash
ko register #fort-nix     # register current project as "fort-nix"
ko register #exo           # register another project
ko default #exo            # set default for unrecognized tags
ko projects                # list all registered projects
```

Registry lives at `~/.config/knockout/projects.yml`.

## Pipeline Configuration

Pipeline config lives in `.ko/pipeline.yml`, prompts in `.ko/prompts/`.
See `examples/` for starter templates.

```yaml
# .ko/pipeline.yml
model: claude-sonnet-4-5-20250929
max_retries: 2
max_depth: 2
discretion: high

stages:
  - name: triage
    prompt: triage.md
    on_fail: blocked

  - name: implement
    prompt: implement.md

  - name: verify
    run: just test
    on_fail: fail

  - name: review
    prompt: review.md
    on_fail: fail

on_succeed:
  - git add -A
  - git commit -m "ko: implement ${TICKET_ID}"
  - git push
```

## Data Model

Tickets are markdown files with YAML frontmatter in `.tickets/`. No database,
no index, no derived state. The file is the source of truth.
