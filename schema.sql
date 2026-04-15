-- Knockout SQLite Schema v1
-- Shadow database for query acceleration. Filesystem remains authoritative.

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER PRIMARY KEY,
    tag         TEXT    NOT NULL UNIQUE,
    prefix      TEXT    NOT NULL,
    tickets_dir TEXT    NOT NULL UNIQUE,
    is_default  INTEGER NOT NULL DEFAULT 0 CHECK (is_default IN (0, 1)),
    created_at  TEXT    NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_single_default
    ON projects(is_default) WHERE is_default = 1;

CREATE TABLE IF NOT EXISTS tickets (
    id           TEXT    PRIMARY KEY,
    ticket_id    TEXT    NOT NULL,
    project_id   INTEGER NOT NULL REFERENCES projects(id),
    title        TEXT    NOT NULL,
    body         TEXT    NOT NULL DEFAULT '',
    status       TEXT    NOT NULL DEFAULT 'open',
    type         TEXT    NOT NULL DEFAULT 'task',
    priority     INTEGER NOT NULL DEFAULT 2,
    assignee     TEXT,
    parent_id    TEXT    REFERENCES tickets(id),
    external_ref TEXT,
    snooze       TEXT,
    triage       TEXT,
    created_at   TEXT    NOT NULL,
    updated_at   TEXT    NOT NULL,

    CONSTRAINT valid_status CHECK (status IN (
        'captured', 'routed', 'open', 'in_progress',
        'closed', 'blocked', 'resolved'
    ))
);

CREATE INDEX IF NOT EXISTS idx_tickets_ticket_id        ON tickets(ticket_id);
CREATE INDEX IF NOT EXISTS idx_tickets_project_status   ON tickets(project_id, status);
CREATE INDEX IF NOT EXISTS idx_tickets_status           ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority_updated ON tickets(priority, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_parent           ON tickets(parent_id);
CREATE INDEX IF NOT EXISTS idx_tickets_snooze           ON tickets(snooze) WHERE snooze IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tickets_triage           ON tickets(triage)  WHERE triage  IS NOT NULL;

CREATE TABLE IF NOT EXISTS ticket_tags (
    ticket_id  TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (ticket_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_ticket_tags_tag ON ticket_tags(tag);

CREATE TABLE IF NOT EXISTS ticket_deps (
    ticket_id  TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    depends_on TEXT NOT NULL,
    PRIMARY KEY (ticket_id, depends_on)
);

CREATE INDEX IF NOT EXISTS idx_ticket_deps_depends_on ON ticket_deps(depends_on);

CREATE TABLE IF NOT EXISTS plan_questions (
    id          INTEGER PRIMARY KEY,
    ticket_id   TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL,
    question    TEXT NOT NULL,
    context     TEXT,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    UNIQUE (ticket_id, question_id)
);

CREATE TABLE IF NOT EXISTS plan_question_options (
    id               INTEGER PRIMARY KEY,
    plan_question_id INTEGER NOT NULL REFERENCES plan_questions(id) ON DELETE CASCADE,
    label            TEXT    NOT NULL,
    value            TEXT    NOT NULL,
    description      TEXT,
    sort_order       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ticket_notes (
    id         INTEGER PRIMARY KEY,
    ticket_id  TEXT    NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    noted_at   TEXT    NOT NULL,
    author     TEXT,
    body       TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ticket_notes_ticket   ON ticket_notes(ticket_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_ticket_notes_noted_at ON ticket_notes(noted_at);

CREATE TABLE IF NOT EXISTS builds (
    id           INTEGER PRIMARY KEY,
    ticket_id    TEXT    NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    workflow     TEXT    NOT NULL DEFAULT 'main',
    started_at   TEXT    NOT NULL,
    completed_at TEXT,
    outcome      TEXT,

    CONSTRAINT valid_outcome CHECK (outcome IN ('succeed', 'fail'))
);

CREATE INDEX IF NOT EXISTS idx_builds_ticket     ON builds(ticket_id);
CREATE INDEX IF NOT EXISTS idx_builds_started_at ON builds(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_builds_outcome    ON builds(outcome) WHERE outcome IS NOT NULL;

CREATE TABLE IF NOT EXISTS build_nodes (
    id           INTEGER PRIMARY KEY,
    build_id     INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    node_name    TEXT    NOT NULL,
    node_type    TEXT    NOT NULL,
    attempt      INTEGER NOT NULL DEFAULT 1,
    started_at   TEXT    NOT NULL,
    completed_at TEXT,
    result       TEXT,
    reason       TEXT,
    sort_order   INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_build_nodes_build ON build_nodes(build_id, sort_order);

CREATE TABLE IF NOT EXISTS build_events (
    id          INTEGER PRIMARY KEY,
    build_id    INTEGER REFERENCES builds(id) ON DELETE CASCADE,
    ticket_id   TEXT    NOT NULL,
    event_type  TEXT    NOT NULL,
    occurred_at TEXT    NOT NULL,
    payload     TEXT
);

CREATE INDEX IF NOT EXISTS idx_build_events_ticket   ON build_events(ticket_id);
CREATE INDEX IF NOT EXISTS idx_build_events_build    ON build_events(build_id) WHERE build_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_build_events_occurred ON build_events(occurred_at DESC);

CREATE TABLE IF NOT EXISTS build_artifacts (
    id            INTEGER PRIMARY KEY,
    build_id      INTEGER NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    node_name     TEXT,
    artifact_path TEXT    NOT NULL,
    content       TEXT,
    created_at    TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_build_artifacts_build ON build_artifacts(build_id);

CREATE TABLE IF NOT EXISTS mutation_events (
    id          INTEGER PRIMARY KEY,
    occurred_at TEXT    NOT NULL,
    project_tag TEXT,
    ticket_id   TEXT,
    event_type  TEXT    NOT NULL,
    payload     TEXT
);

CREATE INDEX IF NOT EXISTS idx_mutation_events_ticket   ON mutation_events(ticket_id) WHERE ticket_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mutation_events_type     ON mutation_events(event_type);
CREATE INDEX IF NOT EXISTS idx_mutation_events_occurred ON mutation_events(occurred_at DESC);

-- Ready queue: what the agent should pick up next
CREATE VIEW IF NOT EXISTS ready_tickets AS
SELECT t.*
FROM tickets t
WHERE t.status IN ('open', 'in_progress')
  AND t.triage IS NULL
  AND (t.snooze IS NULL OR t.snooze < date('now'))
  AND NOT EXISTS (
      SELECT 1
      FROM ticket_deps d
      JOIN tickets dep ON dep.ticket_id = d.depends_on
      WHERE d.ticket_id = t.id
        AND dep.status NOT IN ('closed', 'resolved')
  );

-- Full ticket tree (recursive parent -> children)
CREATE VIEW IF NOT EXISTS ticket_tree AS
WITH RECURSIVE tree(id, ticket_id, parent_id, title, status, depth) AS (
    SELECT id, ticket_id, parent_id, title, status, 0
    FROM tickets
    WHERE parent_id IS NULL
    UNION ALL
    SELECT t.id, t.ticket_id, t.parent_id, t.title, t.status, tree.depth + 1
    FROM tickets t
    JOIN tree ON t.parent_id = tree.id
)
SELECT * FROM tree;

-- Build history summary per ticket
CREATE VIEW IF NOT EXISTS ticket_build_summary AS
SELECT
    ticket_id,
    COUNT(*)                                              AS total_builds,
    SUM(CASE WHEN outcome = 'succeed' THEN 1 ELSE 0 END) AS succeeded,
    SUM(CASE WHEN outcome = 'fail'    THEN 1 ELSE 0 END) AS failed,
    MAX(started_at)                                       AS last_build_at,
    MAX(CASE WHEN outcome = 'succeed' THEN completed_at END) AS last_succeed_at
FROM builds
GROUP BY ticket_id;
