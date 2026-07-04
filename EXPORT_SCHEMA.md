# Knockout Export Schema

`ko export` produces a complete JSON dump of every ticket across every
registered project. This document is the **contract** between ko and the
Questbook bulk-import endpoint (built in a parallel dispatch). The two sides
coordinate on *this file*, not on a shared library or live API. Bump
`schema_version` on any breaking change to the shapes below.

- **Producer:** `ko export` (see `cmd_export.go`, `ExportSchemaVersion`).
- **Consumer:** Questbook QQL bulk-import (`~/Projects/questbook`).
- **Current version:** `1`

## Usage

```
ko export [--out FILE] [--project TAG] [--no-history]
```

- No `--out` (or `--out -`) writes JSON to stdout.
- `--out FILE` writes to a file and prints a one-line summary to stderr.
- `--project TAG` limits the dump to a single project (default: all).
- `--no-history` omits the per-ticket mutation history (smaller, faster).

The command reads the local ko SQLite store directly; it never proxies to a
remote `ko serve`.

## Top level

```json
{
  "schema_version": "1",
  "generator": "ko export <version>",
  "exported_at": "2026-07-04T04:49:15Z",   // RFC3339, UTC
  "project_count": 27,
  "ticket_count": 1405,
  "projects": [ /* ExportProject */ ]
}
```

## ExportProject

```json
{
  "tag": "fort-nix",                 // ko registry tag (the #tag)
  "prefix": "fn",                    // ticket ID prefix
  "path": "/home/dev/Projects/fort-nix",
  "is_default": false,
  "hidden": false,                   // omitted when false
  "tickets": [ /* ExportTicket */ ]
}
```

The importer maps `tag` → realm/campaign via the mapping file (see
`QQL_MAPPING.md`), which shares its shape with this export's `tag` keys.

## ExportTicket

```json
{
  "id": "fn-a001",                   // globally unique in ko
  "title": "…",
  "body": "…",                       // markdown body (may be empty)
  "status": "open",                  // see status vocabulary below
  "type": "task",
  "priority": 2,                     // 0 (highest) … 4 (lowest); ko default 2
  "assignee": "…",                   // omitted when empty
  "parent": "fn-a000",               // omitted when empty
  "external_ref": "…",               // omitted when empty
  "snooze": "2026-05-01",            // ISO date, omitted when empty
  "triage": "…",                     // free-text triage note, omitted when empty
  "deps": ["fn-x", "fn-y"],          // ALWAYS present (may be empty)
  "tags": ["infra"],                 // ALWAYS present (may be empty)
  "plan_questions": [ /* PlanQuestion */ ],  // omitted when none
  "created": "2026-02-26T01:24:41Z", // RFC3339
  "modified": "2026-04-07T03:41:58Z",// RFC3339, from the store's updated_at
  "history": [ /* ExportEvent */ ]   // omitted when none / --no-history
}
```

**Contract guarantees for importers:**

- `deps` and `tags` are **always arrays**, never `null`. Everything else marked
  "omitted when empty" uses `omitempty` and may be absent.
- `id` is globally unique across all projects, so cross-project `deps` and
  `parent` references resolve unambiguously.
- `deps` are ticket IDs this ticket depends on (blockers). In QQL terms, a dep
  `A → B` means quest A `dependencies+` B.

### Status vocabulary (ko → QQL)

ko statuses and the QQL quest statuses they map to (see `koToQQLStatus` in
`cmd_shim.go`; the importer should apply the same mapping):

| ko status     | QQL quest status |
|---------------|------------------|
| `captured`    | `open`           |
| `routed`      | `open`           |
| `open`        | `open`           |
| `in_progress` | `in_progress`    |
| `blocked`     | `blocked`        |
| `closed`      | `done`           |
| `resolved`    | `done`           |

### PlanQuestion

ko's per-ticket disambiguation questions (the Punchlist pattern). These map
onto Questbook's suggestion/question surface (Slice 3).

```json
{
  "id": "q1",
  "question": "…",
  "context": "…",                    // omitted when empty
  "options": [
    { "label": "…", "value": "…", "description": "…" }
  ]
}
```

### ExportEvent

One entry of a ticket's mutation history, oldest first. Cheap to include (a
single bulk query grouped by ticket ID).

```json
{
  "occurred_at": "2026-06-28T06:20:13Z",  // RFC3339
  "event_type": "create",                 // create | update | dep | undep | note | …
  "payload": { "title": "…" }             // raw JSON recorded with the event; omitted when none
}
```

`payload` is passed through verbatim as recorded by ko; its shape varies by
`event_type` and is advisory (provenance), not a stable sub-contract.
