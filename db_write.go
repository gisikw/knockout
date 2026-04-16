package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ensureProject finds or creates a project row for the given tickets directory.
// Returns the project's integer ID.
func (d *DB) ensureProject(ticketsDir string) (int64, error) {
	abs, err := filepath.Abs(ticketsDir)
	if err != nil {
		abs = ticketsDir
	}

	var id int64
	err = d.db.QueryRow("SELECT id FROM projects WHERE tickets_dir = ?", abs).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Derive prefix and tag from the tickets directory.
	prefix := detectPrefixFromDir(abs)
	if prefix == "" {
		prefix = "unknown"
	}
	// Tag: use the directory two levels up as a reasonable default.
	tag := filepath.Base(ProjectRoot(abs))

	res, err := d.db.Exec(
		"INSERT INTO projects (tag, prefix, tickets_dir, created_at) VALUES (?, ?, ?, ?)",
		tag, prefix, abs, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpsertTicket writes a ticket and its relations to the shadow database.
// All writes happen in a single transaction.
func (d *DB) UpsertTicket(t *Ticket, ticketsDir string) error {
	projectID, err := d.ensureProject(ticketsDir)
	if err != nil {
		return fmt.Errorf("ensure project: %w", err)
	}

	prefix := extractPrefix(t.ID)
	uuid := ticketUUID(prefix, t.ID)
	now := time.Now().UTC().Format(time.RFC3339)

	// Resolve parent UUID if present, but only set FK if parent exists in DB.
	// Parents may be cross-project or deleted — soft reference is correct.
	var parentUUID *string
	if t.Parent != "" {
		p := ticketUUID(extractPrefix(t.Parent), t.Parent)
		var exists int
		if d.db.QueryRow("SELECT 1 FROM tickets WHERE id = ?", p).Scan(&exists) == nil {
			parentUUID = &p
		}
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert ticket row.
	_, err = tx.Exec(`
		INSERT INTO tickets (id, ticket_id, project_id, title, body, status, type, priority,
			assignee, parent_id, external_ref, snooze, triage, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title=excluded.title, body=excluded.body, status=excluded.status,
			type=excluded.type, priority=excluded.priority, assignee=excluded.assignee,
			parent_id=excluded.parent_id, external_ref=excluded.external_ref,
			snooze=excluded.snooze, triage=excluded.triage, updated_at=excluded.updated_at`,
		uuid, t.ID, projectID, t.Title, t.Body, t.Status, t.Type, t.Priority,
		nullStr(t.Assignee), parentUUID, nullStr(t.ExternalRef),
		nullStr(t.Snooze), nullStr(t.Triage),
		coalesceStr(t.Created, now), now,
	)
	if err != nil {
		return fmt.Errorf("upsert ticket: %w", err)
	}

	// Replace tags.
	if _, err := tx.Exec("DELETE FROM ticket_tags WHERE ticket_id = ?", uuid); err != nil {
		return err
	}
	for _, tag := range t.Tags {
		if _, err := tx.Exec("INSERT INTO ticket_tags (ticket_id, tag) VALUES (?, ?)", uuid, tag); err != nil {
			return err
		}
	}

	// Replace deps.
	if _, err := tx.Exec("DELETE FROM ticket_deps WHERE ticket_id = ?", uuid); err != nil {
		return err
	}
	for _, dep := range t.Deps {
		if _, err := tx.Exec("INSERT INTO ticket_deps (ticket_id, depends_on) VALUES (?, ?)", uuid, dep); err != nil {
			return err
		}
	}

	// Replace plan questions + options.
	// Delete options first (cascade would handle it, but explicit is safer with soft deletes).
	if _, err := tx.Exec(`DELETE FROM plan_question_options WHERE plan_question_id IN
		(SELECT id FROM plan_questions WHERE ticket_id = ?)`, uuid); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM plan_questions WHERE ticket_id = ?", uuid); err != nil {
		return err
	}
	for i, q := range t.PlanQuestions {
		res, err := tx.Exec(
			"INSERT INTO plan_questions (ticket_id, question_id, question, context, sort_order) VALUES (?, ?, ?, ?, ?)",
			uuid, q.ID, q.Question, nullStr(q.Context), i,
		)
		if err != nil {
			return err
		}
		qID, _ := res.LastInsertId()
		for j, opt := range q.Options {
			if _, err := tx.Exec(
				"INSERT INTO plan_question_options (plan_question_id, label, value, description, sort_order) VALUES (?, ?, ?, ?, ?)",
				qID, opt.Label, opt.Value, nullStr(opt.Description), j,
			); err != nil {
				return err
			}
		}
	}

	// Replace notes (parsed from body).
	if _, err := tx.Exec("DELETE FROM ticket_notes WHERE ticket_id = ?", uuid); err != nil {
		return err
	}
	for i, note := range parseNotes(t.Body) {
		if _, err := tx.Exec(
			"INSERT INTO ticket_notes (ticket_id, noted_at, author, body, sort_order) VALUES (?, ?, ?, ?, ?)",
			uuid, note.notedAt, nullStr(note.author), note.body, i,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpsertProject writes a project row to the shadow database.
func (d *DB) UpsertProject(tag, prefix, ticketsDir string, isDefault bool) error {
	abs, err := filepath.Abs(ticketsDir)
	if err != nil {
		abs = ticketsDir
	}
	now := time.Now().UTC().Format(time.RFC3339)
	def := 0
	if isDefault {
		def = 1
	}
	_, err = d.db.Exec(`
		INSERT INTO projects (tag, prefix, tickets_dir, is_default, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(tag) DO UPDATE SET
			prefix=excluded.prefix, tickets_dir=excluded.tickets_dir,
			is_default=excluded.is_default`,
		tag, prefix, abs, def, now,
	)
	return err
}

// InsertMutationEvent writes a mutation event to the shadow database.
func (d *DB) InsertMutationEvent(e MutationEvent) error {
	var payload *string
	if e.Data != nil {
		b, err := json.Marshal(e.Data)
		if err == nil {
			s := string(b)
			payload = &s
		}
	}
	_, err := d.db.Exec(
		"INSERT INTO mutation_events (occurred_at, project_tag, ticket_id, event_type, payload) VALUES (?, ?, ?, ?, ?)",
		e.Timestamp, nullStr(projectTagFromPath(e.Project)), nullStr(e.Ticket), e.Event, payload,
	)
	return err
}

// InsertBuildEvent writes a raw build event to the shadow database.
func (d *DB) InsertBuildEvent(ticketUUID, eventType, occurredAt, payload string) error {
	_, err := d.db.Exec(
		"INSERT INTO build_events (ticket_id, event_type, occurred_at, payload) VALUES (?, ?, ?, ?)",
		ticketUUID, eventType, occurredAt, payload,
	)
	return err
}

// SyncRegistry writes all projects from a Registry to the shadow database.
func (d *DB) SyncRegistry(reg *Registry) error {
	for tag, path := range reg.Projects {
		prefix := reg.Prefixes[tag]
		if prefix == "" {
			prefix = tag
		}
		isDefault := reg.Default == tag
		ticketsDir := resolveTicketsDir(path)
		if err := d.UpsertProject(tag, prefix, ticketsDir, isDefault); err != nil {
			return err
		}
	}
	return nil
}

// parsedNote is a note extracted from the ticket body's ## Notes section.
type parsedNote struct {
	notedAt string
	author  string
	body    string
}

// noteRe matches "**2026-01-15 10:30:45 UTC:** some text"
var noteRe = regexp.MustCompile(`\*\*(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} UTC):\*\* (.*)`)

// parseNotes extracts structured notes from a ticket body.
func parseNotes(body string) []parsedNote {
	idx := strings.Index(body, "## Notes")
	if idx < 0 {
		return nil
	}
	section := body[idx:]
	matches := noteRe.FindAllStringSubmatch(section, -1)
	notes := make([]parsedNote, 0, len(matches))
	for _, m := range matches {
		ts := m[1]
		text := m[2]
		author := ""
		// Check for author prefix (ko: | agent: | operator:)
		for _, prefix := range []string{"ko:", "agent:", "operator:"} {
			if strings.HasPrefix(text, prefix+" ") {
				author = strings.TrimSuffix(prefix, ":")
				text = strings.TrimPrefix(text, prefix+" ")
				break
			}
		}
		notes = append(notes, parsedNote{notedAt: ts, author: author, body: text})
	}
	return notes
}

// nullStr returns nil for empty strings, otherwise a pointer to s.
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// coalesceStr returns a if non-empty, otherwise b.
func coalesceStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// projectTagFromPath derives a project tag from an absolute project path.
func projectTagFromPath(projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	return filepath.Base(projectRoot)
}

// writeTicketToDB writes a ticket to SQLite as the authoritative store.
// Returns error if write fails. DB is required for production use.
func writeTicketToDB(t *Ticket, ticketsDir string) error {
	db := getShadowDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.UpsertTicket(t, ticketsDir)
}

// isTempDir returns true if path is under a temp directory.
// Checks both /tmp and os.TempDir() since TMPDIR env var may differ.
func isTempDir(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	// Check standard /tmp
	if strings.HasPrefix(abs, "/tmp/") {
		return true
	}
	// Check TMPDIR (may differ in some environments)
	tmpDir := os.TempDir()
	return strings.HasPrefix(abs, tmpDir)
}

// shadowWriteTicket is the best-effort shadow write called after SaveTicket.
// Deprecated: Use writeTicketToDB for authoritative writes.
func shadowWriteTicket(t *Ticket, ticketsDir string) {
	db := getShadowDB()
	if db == nil {
		return
	}
	if err := db.UpsertTicket(t, ticketsDir); err != nil {
		fmt.Fprintf(os.Stderr, "ko: shadow write ticket %s: %v\n", t.ID, err)
	}
}

// shadowWriteRegistry is the best-effort shadow write called after SaveRegistry.
func shadowWriteRegistry(reg *Registry) {
	db := getShadowDB()
	if db == nil {
		return
	}
	if err := db.SyncRegistry(reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko: shadow write registry: %v\n", err)
	}
}

// shadowWriteMutation is the best-effort shadow write called after EmitMutationEvent.
func shadowWriteMutation(e MutationEvent) {
	db := getShadowDB()
	if db == nil {
		return
	}
	if err := db.InsertMutationEvent(e); err != nil {
		fmt.Fprintf(os.Stderr, "ko: shadow write mutation: %v\n", err)
	}
}
