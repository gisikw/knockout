# knockout

A task tracker and build pipeline for autonomous agent workflows. Track tickets,
route work across projects, and burn down backlogs with BFS over the ready queue.

## What It Does

`ko` is a single binary that combines ticket management with an automated build
pipeline. Every `ko build` outcome mutates the ready queue — tickets get closed,
blocked, or decomposed into subtasks. Nothing stays unchanged after a build
attempt.

## Quick Start

```bash
# Create a ticket
ko create "Fix the auth bug"

# See what's ready to work on
ko ready

# Build it (run the automated pipeline)
ko build ko-a1b2

# Capture something for another project
ko add "Add foobar to dev-sandbox #fort-nix"
```

## Concepts

### Statuses

| Status | Meaning |
|--------|---------|
| `captured` | Just captured, probably doesn't belong here yet |
| `routed` | Triaged to this repo, needs evaluation |
| `open` | Ready to be worked (eligible for `ko ready`) |
| `in_progress` | Currently being worked |
| `closed` | Done |
| `blocked` | Needs human attention |

`ko ready` only returns `open` and `in_progress` tickets with all deps resolved.

### Hierarchical IDs

Ticket IDs encode parent-child relationships:

```
ko-a1b2          # root ticket (depth 0)
ko-a1b2.c3d4     # child (depth 1)
ko-a1b2.c3d4.e5f6  # grandchild (depth 2)
```

Depth is visible by counting dots. Decomposition is bounded by max depth —
at the limit, tickets get blocked for human review instead of further decomposed.

### Build Pipeline

`ko build <ticket-id>` runs a YAML-defined pipeline against a ticket. Every
outcome removes the ticket from the ready queue:

| Outcome | Effect |
|---------|--------|
| `SUCCEED` | Ticket closed. Queue shrinks. |
| `FAIL` | Ticket blocked (needs human). Reason note explains why. |
| `BLOCKED` | Dependency wired. Ticket off queue until dep resolves. |
| `DECOMPOSE` | Child tickets created, parent blocked on them. Queue reshapes. |

### Cross-Project Routing

`ko add "thing #project-name"` captures a task and routes it to the target
project via the project registry. Unrecognized tags go to the default project
as a catch-all — repeated unrecognized tags signal that a project wants to exist.

Cross-project deps are resolved lazily: `ko ready` only checks them when the
local queue is empty, and short-circuits on the first unblocked ticket.

### Pipeline Configuration

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
  - echo "${CHANGED_FILES}" | xargs git add
  - git add .tickets/${TICKET_ID}.md
  - git commit -m "ko: implement ${TICKET_ID}"
  - git push

on_close:
  - just deploy
```

## Data Model

Tickets are markdown files with YAML frontmatter in `.tickets/`. No database,
no index, no derived state. The file is the source of truth.

## Status

Pre-implementation. Specs and invariants are defined. The reference implementation
is [tk-build](https://git.gisi.network/infra/tk-build) (bash), which this
project supersedes.
