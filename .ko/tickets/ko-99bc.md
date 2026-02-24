---
id: ko-99bc
status: closed
deps: []
links: []
created: 2026-02-23T02:12:56Z
type: task
priority: 2
---
# Add JSONL backmatter to ticket files for structured event history

## Design

Append structured JSONL event history below a `---` separator in ticket .md files.

### Format

```
---
{"ts":"2026-02-22T06:38:21Z","ev":"build","attempt":1,"node":"triage","disposition":"continue"}
{"ts":"2026-02-22T06:38:45Z","ev":"build","attempt":1,"node":"verify","disposition":"fail","reason":"cargo test E0761"}
{"ts":"2026-02-22T06:39:16Z","ev":"status","to":"blocked","by":"agent","reason":"Rust module conflict blocks verify"}
{"ts":"2026-02-22T21:47:00Z","ev":"status","to":"closed","by":"human","note":"confirmed duplicate"}
```

### Event types

- `ev:status` — status transitions with `to`, `from`, `by` (agent|human), `reason`
- `ev:build` — pipeline node results with `attempt`, `node`, `disposition`, `reason`

### Principles

- Append-only, tail-friendly (`tail -n3` gives you the recent story)
- Machine events only (status changes, triage dispositions, build results) — not a general activity log
- `reason` field does the heavy lifting for after-action reports
- `by` field distinguishes agent vs human actions (critical for reconstructing overnight runs)
- Markdown body stays for humans; backmatter is for agents

### Enables

- `ko agent report` — after-action summary of autonomous runs (failures, questions for human, attempt counts)
- Cold-start context — new sessions can read ticket history without hunting through .ko/builds/
- Pattern detection — "this ticket failed 3x at verify" vs "failed once at triage"
