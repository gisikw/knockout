# WORKLOG — Knockout → Questbook compat shim + export

Dispatch: `instructions-qql-shim.md`. Builds the *leaving* side of the Knockout
→ Questbook migration: a full export, a `ko`-shaped compat shim over QQL, and a
read-only legacy mode. Design authority: `~/briefs/questbook.md` §10 Slice 1.

## What shipped

| Piece | Files | Opt-in |
|-------|-------|--------|
| **Export** | `cmd_export.go`, `db_read.go` (`ExportTicketHistory`) | always available: `ko export` |
| **QQL client** | `qql_client.go` | — |
| **Mapping** | `qql_mapping.go`, `qql-mapping.example.yaml` | file, `KO_QQL_MAPPING` |
| **Shim** | `cmd_shim.go` | `KO_QQL=1` (inert otherwise) |
| **Read-only legacy** | `readonly.go` | `KO_READONLY=1` (inert otherwise) |
| **Wiring** | `main.go`, `remote.go` | — |
| **Docs** | `EXPORT_SCHEMA.md`, `QQL_MAPPING.md`, this file | — |
| **Tests** | `cmd_export_test.go`, `cmd_shim_test.go`, `qql_mapping_test.go`, `readonly_test.go` | — |

All three toggles are independent and **off by default** — nothing about
default `ko` behavior changes until Kevin flips a switch. The flip is his
coordination moment, per the dispatch boundary.

## Acceptance status

1. ✅ Export produces a complete dump (verified: 27 projects, 1405 tickets);
   schema documented in `EXPORT_SCHEMA.md`.
2. ✅ Shim round-trips against **live** qb.gisi.network: `add`/`ls`/`show`/`dep`
   create real entities (quests `q-b9027c66`, `q-df5991c0` in realm
   `knockout-shim-smoke`), plus `start`/`close`/`ready`.
3. ✅ Every shim invocation logged as JSONL
   (`~/.local/state/knockout/shim-usage.jsonl`, override `KO_SHIM_LOG`).
4. ✅ Unsupported subcommands fail loudly with a QQL pointer (exit 2), never
   silently no-op.
5. ✅ Read-only mode rejects writes (exit 3) with a pointer, allows reads.
6. ✅ Existing tests green (`go test ./...`), plus 19 new tests; this WORKLOG.

## Key decisions

- **Opt-in via env var, not a separate binary.** `KO_QQL=1` routes at the top of
  `run()`. Simpler than a second build target and trivially inert. `KO_READONLY=1`
  is a separate, independent gate on the *legacy* store.
- **Realm is the anchor; campaigns are never auto-created.** Realms need only a
  slug+name (shim creates them on demand). Campaigns require a mandatory `goal`
  we don't have at `ko add` time, so the shim attaches a campaign only if its
  slug already exists, else warns and anchors to the realm. This satisfies the
  QQL quest invariant (`realm ∨ campaign ∨ parent`) without fabricating goals.
- **New IDs, not preserved ko IDs.** The shim is a *live* surface: `ko add`
  creates a quest and prints its `q-…` id. Agent habits capture the printed id
  and reuse it — same argv shape, same output shape, different id space. QQL has
  no external-ref column, so ko IDs aren't round-tripped through the shim (they
  ARE preserved in `ko export` for the bulk importer).
- **Status mapping** ko→QQL is centralized in `koToQQLStatus` and documented in
  `EXPORT_SCHEMA.md` so the importer applies the identical mapping.
- **Export reads the DB directly** and is marked local-only in `remote.go` so it
  never proxies to a remote `ko serve`.
- **History is cheap:** one bulk query of `mutation_events` grouped by ticket id
  (globally unique), attached per ticket. `--no-history` opts out.
- **Boring by design.** No ko redesign; the shim is meant to rot away. The
  usage log going quiet is the evidence migration is done.

## Parked questions / known limitations

- **QQL query cannot read quest relations (the big one).** The deployed
  `Store.Query.project()` drops every dotted relation projection
  (`dependencies.*`, `subquests.*`, `parent.slug`) — only scalar columns come
  back, with realm/campaign/parent as bare IDs. Consequences, all handled
  *honestly* rather than faked:
  - `ko show` prints realm/campaign/parent as IDs and shows **no deps line**
    (it can't read them). Dependencies you write with `ko dep` DO persist
    (the mutate is transactional and would error on a bad target) — they're
    just not queryable back yet.
  - `ko ready` degrades to "open quests in realm" (a superset of the true ready
    queue) and **prints a stderr warning** so it's never a silent lie. Wire real
    dep-gating once QQL exposes dependency status in queries.
  - `ko dep tree` **fails loudly** — a dependency graph can't be walked without
    relation reads. Better than a misleading single-node tree.
  → *Ask for the Questbook side:* expose quest `dependencies`/`subquests` (at
  least ids + status) in `Query.project()`. That single change lets `show`,
  `ready`, and `dep tree` become faithful.
- **No auth on QQL.** The API currently takes no token; the client sends none.
  Add a header in `qql_client.go` when Questbook grows auth.
- **ko-only fields with no QQL home** are dropped by the shim (assignee, tags,
  snooze, triage, external-ref, plan-questions). They're accepted on the argv so
  muscle memory doesn't break, logged as dropped by omission, and fully
  preserved in `ko export` for the importer to place. Revisit as Questbook's
  schema grows (tags/assignee likely; snooze/triage may stay ko-only).
- **`ready` dep-approximation** is a superset today. If agents drive off `ko
  ready` through the shim before QQL relation-reads land, they may pick up
  dep-blocked quests. The stderr warning flags it; gate for real ASAP.
- **Smoke-test entities** remain in the live `knockout-shim-smoke` realm as
  acceptance evidence (§2). Delete via `qb mutate` if the clutter bothers.

## Verified behaviors (manual, against live API)

```
$ KO_QQL=1 ko add "…"          → prints q-…              (real quest created)
$ KO_QQL=1 ko ls               → id [status] (pN) title  (realm-scoped)
$ KO_QQL=1 ko show q-…         → scalar fields (honest, no faked deps)
$ KO_QQL=1 ko dep q-a q-b      → Added dependency: q-a -> q-b
$ KO_QQL=1 ko start/close q-…  → q-… updated             (status transition)
$ KO_QQL=1 ko stats            → fails loud, exit 2, QQL pointer
$ KO_READONLY=1 ko add "…"     → rejected, exit 3, pointer to shim/QQL
$ KO_READONLY=1 ko ls          → works (read allowed)
$ ko export --out dump.json    → 27 projects / 1405 tickets
```
