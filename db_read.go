package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// OpenReadDB opens the shadow DB for reading. Unlike getShadowDB, this
// returns an actionable error if the DB is missing or empty.
func OpenReadDB() (*DB, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w\n\nRun 'ko import' to populate the database", err)
	}

	// Verify it has data
	var count int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&count); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot query tickets: %w", err)
	}
	if count == 0 {
		db.Close()
		return nil, fmt.Errorf("database is empty\n\nRun 'ko import' to populate the database")
	}

	return db, nil
}

// StatusCount holds a status string and its count.
type StatusCount struct {
	Status string
	Count  int
}

// TypeCount holds a type string and its count.
type TypeCount struct {
	Type  string
	Count int
}

// PriorityCount holds a priority and its count.
type PriorityCount struct {
	Priority int
	Count    int
}

// ProjectStats holds per-project aggregate stats.
type ProjectStats struct {
	Tag    string
	Open   int
	Closed int
}

// StatsResult holds all aggregate statistics.
type StatsResult struct {
	Total           int
	ByStatus        []StatusCount
	ByType          []TypeCount
	ByPriority      []PriorityCount
	CreatedThisWeek int
	ClosedThisWeek  int
	CreatedThisMonth int
	ClosedThisMonth int
	Ready           int
	Blocked         int
	TotalBuilds     int
	Succeeded       int
	Failed          int
	ByProject       []ProjectStats
}

// QueryStats returns aggregate metrics, optionally filtered by project tag.
func (d *DB) QueryStats(project string) (*StatsResult, error) {
	result := &StatsResult{}

	// Build WHERE clause for project filter
	var projectWhere string
	var projectArgs []interface{}
	if project != "" {
		projectWhere = " AND p.tag = ?"
		projectArgs = append(projectArgs, project)
	}

	// Total tickets
	q := "SELECT COUNT(*) FROM tickets t JOIN projects p ON t.project_id = p.id WHERE 1=1" + projectWhere
	if err := d.db.QueryRow(q, projectArgs...).Scan(&result.Total); err != nil {
		return nil, err
	}

	// By status
	q = `SELECT status, COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE 1=1` + projectWhere + ` GROUP BY status ORDER BY COUNT(*) DESC`
	rows, err := d.db.Query(q, projectArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		result.ByStatus = append(result.ByStatus, sc)
	}

	// By type
	q = `SELECT type, COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE 1=1` + projectWhere + ` GROUP BY type ORDER BY COUNT(*) DESC`
	rows, err = d.db.Query(q, projectArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tc TypeCount
		if err := rows.Scan(&tc.Type, &tc.Count); err != nil {
			return nil, err
		}
		result.ByType = append(result.ByType, tc)
	}

	// By priority
	q = `SELECT priority, COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE 1=1` + projectWhere + ` GROUP BY priority ORDER BY priority`
	rows, err = d.db.Query(q, projectArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pc PriorityCount
		if err := rows.Scan(&pc.Priority, &pc.Count); err != nil {
			return nil, err
		}
		result.ByPriority = append(result.ByPriority, pc)
	}

	// Created/closed this week and month
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

	q = `SELECT COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE created_at >= ?` + projectWhere
	args := append([]interface{}{weekAgo}, projectArgs...)
	d.db.QueryRow(q, args...).Scan(&result.CreatedThisWeek)

	q = `SELECT COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE status = 'closed' AND updated_at >= ?` + projectWhere
	d.db.QueryRow(q, args...).Scan(&result.ClosedThisWeek)

	args = append([]interface{}{monthAgo}, projectArgs...)
	q = `SELECT COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE created_at >= ?` + projectWhere
	d.db.QueryRow(q, args...).Scan(&result.CreatedThisMonth)

	q = `SELECT COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE status = 'closed' AND updated_at >= ?` + projectWhere
	d.db.QueryRow(q, args...).Scan(&result.ClosedThisMonth)

	// Ready count (from view)
	q = `SELECT COUNT(*) FROM ready_tickets r
		 JOIN projects p ON r.project_id = p.id WHERE 1=1` + projectWhere
	d.db.QueryRow(q, projectArgs...).Scan(&result.Ready)

	// Blocked count
	q = `SELECT COUNT(*) FROM tickets t
		 JOIN projects p ON t.project_id = p.id WHERE status = 'blocked'` + projectWhere
	d.db.QueryRow(q, projectArgs...).Scan(&result.Blocked)

	// Build stats
	if project != "" {
		q = `SELECT COUNT(*),
			 SUM(CASE WHEN outcome = 'succeed' THEN 1 ELSE 0 END),
			 SUM(CASE WHEN outcome = 'fail' THEN 1 ELSE 0 END)
			 FROM builds b
			 JOIN tickets t ON b.ticket_id = t.id
			 JOIN projects p ON t.project_id = p.id
			 WHERE p.tag = ?`
		d.db.QueryRow(q, project).Scan(&result.TotalBuilds, &result.Succeeded, &result.Failed)
	} else {
		q = `SELECT COUNT(*),
			 SUM(CASE WHEN outcome = 'succeed' THEN 1 ELSE 0 END),
			 SUM(CASE WHEN outcome = 'fail' THEN 1 ELSE 0 END)
			 FROM builds`
		d.db.QueryRow(q).Scan(&result.TotalBuilds, &result.Succeeded, &result.Failed)
	}

	// Per-project breakdown (only when not filtering by project)
	if project == "" {
		q = `SELECT p.tag,
			 SUM(CASE WHEN t.status NOT IN ('closed', 'resolved') THEN 1 ELSE 0 END) as open,
			 SUM(CASE WHEN t.status IN ('closed', 'resolved') THEN 1 ELSE 0 END) as closed
			 FROM tickets t
			 JOIN projects p ON t.project_id = p.id
			 GROUP BY p.tag
			 ORDER BY open DESC`
		rows, err = d.db.Query(q)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var ps ProjectStats
			if err := rows.Scan(&ps.Tag, &ps.Open, &ps.Closed); err != nil {
				return nil, err
			}
			result.ByProject = append(result.ByProject, ps)
		}
	}

	return result, nil
}

// SearchResult holds a single search result.
type SearchResult struct {
	TicketID string
	Project  string
	Status   string
	Priority int
	Title    string
	Snippet  string
}

// SearchTickets searches tickets by title and body with LIKE.
// Multiple words are ANDed together.
func (d *DB) SearchTickets(query, project, status, ticketType, tag string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	words := strings.Fields(query)
	if len(words) == 0 {
		return nil, fmt.Errorf("search query is empty")
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	// Each word must appear in title OR body
	for _, w := range words {
		pattern := "%" + w + "%"
		conditions = append(conditions, "(t.title LIKE ? OR t.body LIKE ?)")
		args = append(args, pattern, pattern)
	}

	if project != "" {
		conditions = append(conditions, "p.tag = ?")
		args = append(args, project)
	}
	if status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, status)
	}
	if ticketType != "" {
		conditions = append(conditions, "t.type = ?")
		args = append(args, ticketType)
	}
	if tag != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM ticket_tags tt WHERE tt.ticket_id = t.id AND tt.tag = ?)")
		args = append(args, tag)
	}

	q := `SELECT t.ticket_id, p.tag, t.status, t.priority, t.title, t.body
		  FROM tickets t
		  JOIN projects p ON t.project_id = p.id
		  WHERE ` + strings.Join(conditions, " AND ") + `
		  ORDER BY t.priority, t.updated_at DESC
		  LIMIT ?`
	args = append(args, limit)

	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var body string
		if err := rows.Scan(&r.TicketID, &r.Project, &r.Status, &r.Priority, &r.Title, &body); err != nil {
			return nil, err
		}
		// Extract snippet around first match in body
		r.Snippet = extractSnippet(body, words[0], 60)
		results = append(results, r)
	}

	return results, nil
}

// extractSnippet finds the first occurrence of term in text and returns
// surrounding context up to maxLen chars.
func extractSnippet(text, term string, maxLen int) string {
	lower := strings.ToLower(text)
	termLower := strings.ToLower(term)
	idx := strings.Index(lower, termLower)
	if idx < 0 {
		return ""
	}

	// Find start and end of snippet
	start := idx - maxLen/2
	if start < 0 {
		start = 0
	}
	end := idx + len(term) + maxLen/2
	if end > len(text) {
		end = len(text)
	}

	snippet := text[start:end]
	snippet = strings.TrimSpace(snippet)
	snippet = strings.ReplaceAll(snippet, "\n", " ")

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}

// BuildEntry holds build history for display.
type BuildEntry struct {
	TicketID    string
	Project     string
	Workflow    string
	StartedAt   string
	CompletedAt sql.NullString
	Outcome     sql.NullString
	Duration    string
}

// QueryTicketBuilds returns builds for a specific ticket.
func (d *DB) QueryTicketBuilds(ticketID string, limit int) ([]BuildEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	q := `SELECT t.ticket_id, p.tag, b.workflow, b.started_at, b.completed_at, b.outcome
		  FROM builds b
		  JOIN tickets t ON b.ticket_id = t.id
		  JOIN projects p ON t.project_id = p.id
		  WHERE t.ticket_id = ?
		  ORDER BY b.started_at DESC
		  LIMIT ?`

	rows, err := d.db.Query(q, ticketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBuilds(rows)
}

// QueryRecentBuilds returns recent builds across all or one project.
func (d *DB) QueryRecentBuilds(project string, limit int) ([]BuildEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	var q string
	var args []interface{}

	if project != "" {
		q = `SELECT t.ticket_id, p.tag, b.workflow, b.started_at, b.completed_at, b.outcome
			 FROM builds b
			 JOIN tickets t ON b.ticket_id = t.id
			 JOIN projects p ON t.project_id = p.id
			 WHERE p.tag = ?
			 ORDER BY b.started_at DESC
			 LIMIT ?`
		args = []interface{}{project, limit}
	} else {
		q = `SELECT t.ticket_id, p.tag, b.workflow, b.started_at, b.completed_at, b.outcome
			 FROM builds b
			 JOIN tickets t ON b.ticket_id = t.id
			 JOIN projects p ON t.project_id = p.id
			 ORDER BY b.started_at DESC
			 LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBuilds(rows)
}

func scanBuilds(rows *sql.Rows) ([]BuildEntry, error) {
	var results []BuildEntry
	for rows.Next() {
		var b BuildEntry
		if err := rows.Scan(&b.TicketID, &b.Project, &b.Workflow, &b.StartedAt, &b.CompletedAt, &b.Outcome); err != nil {
			return nil, err
		}
		// Calculate duration if completed
		if b.CompletedAt.Valid && b.StartedAt != "" {
			start, err1 := time.Parse(time.RFC3339, b.StartedAt)
			end, err2 := time.Parse(time.RFC3339, b.CompletedAt.String)
			if err1 == nil && err2 == nil {
				dur := end.Sub(start)
				if dur >= time.Minute {
					b.Duration = fmt.Sprintf("%dm %02ds", int(dur.Minutes()), int(dur.Seconds())%60)
				} else {
					b.Duration = fmt.Sprintf("%ds", int(dur.Seconds()))
				}
			}
		}
		results = append(results, b)
	}
	return results, nil
}

// MutationEntry holds mutation event history.
type MutationEntry struct {
	OccurredAt string
	Project    string
	TicketID   string
	EventType  string
	Payload    sql.NullString
}

// QueryTicketMutations returns mutation events for a specific ticket.
func (d *DB) QueryTicketMutations(ticketID string, limit int) ([]MutationEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	q := `SELECT occurred_at, COALESCE(project_tag, ''), COALESCE(ticket_id, ''), event_type, payload
		  FROM mutation_events
		  WHERE ticket_id = ?
		  ORDER BY occurred_at DESC
		  LIMIT ?`

	rows, err := d.db.Query(q, ticketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMutations(rows)
}

// QueryRecentMutations returns recent mutation events across all or one project.
func (d *DB) QueryRecentMutations(project string, limit int) ([]MutationEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	var q string
	var args []interface{}

	if project != "" {
		q = `SELECT occurred_at, COALESCE(project_tag, ''), COALESCE(ticket_id, ''), event_type, payload
			 FROM mutation_events
			 WHERE project_tag = ?
			 ORDER BY occurred_at DESC
			 LIMIT ?`
		args = []interface{}{project, limit}
	} else {
		q = `SELECT occurred_at, COALESCE(project_tag, ''), COALESCE(ticket_id, ''), event_type, payload
			 FROM mutation_events
			 ORDER BY occurred_at DESC
			 LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMutations(rows)
}

func scanMutations(rows *sql.Rows) ([]MutationEntry, error) {
	var results []MutationEntry
	for rows.Next() {
		var m MutationEntry
		if err := rows.Scan(&m.OccurredAt, &m.Project, &m.TicketID, &m.EventType, &m.Payload); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

// GetTicketTitle looks up just the title for a ticket ID.
func (d *DB) GetTicketTitle(ticketID string) (string, error) {
	var title string
	err := d.db.QueryRow("SELECT title FROM tickets WHERE ticket_id = ?", ticketID).Scan(&title)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("ticket not found: %s", ticketID)
	}
	return title, err
}

// ResolveProjectTag returns the project tag for a given tickets directory path.
// Returns empty string if no matching project is found.
func (d *DB) ResolveProjectTag(ticketsDir string) (string, error) {
	if ticketsDir == "" {
		return "", nil
	}
	var tag string
	err := d.db.QueryRow("SELECT tag FROM projects WHERE tickets_dir = ?", ticketsDir).Scan(&tag)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return tag, err
}

// ListTicketsDB returns tickets for ko ls, filtered by project and status.
func (d *DB) ListTicketsDB(project, status string, includeAll bool, limit int) ([]*Ticket, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "1=1")

	if project != "" {
		conditions = append(conditions, "p.tag = ?")
		args = append(args, project)
	}

	if status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, status)
	} else if !includeAll {
		// Default: exclude closed tickets
		conditions = append(conditions, "t.status != 'closed'")
	}

	q := `SELECT t.ticket_id, t.title, t.status, t.type, t.priority,
		         t.assignee, parent.ticket_id, t.external_ref, t.snooze, t.triage,
		         t.created_at, t.updated_at, t.body
		  FROM tickets t
		  JOIN projects p ON t.project_id = p.id
		  LEFT JOIN tickets parent ON t.parent_id = parent.id
		  WHERE ` + strings.Join(conditions, " AND ") + `
		  ORDER BY t.priority, t.updated_at DESC`

	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		t := &Ticket{
			Deps: []string{},
			Tags: []string{},
		}
		var assignee, parentTicketID, extRef, snooze, triage sql.NullString
		var updatedAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.Type, &t.Priority,
			&assignee, &parentTicketID, &extRef, &snooze, &triage,
			&t.Created, &updatedAt, &t.Body); err != nil {
			return nil, err
		}
		t.Assignee = assignee.String
		t.Parent = parentTicketID.String
		t.ExternalRef = extRef.String
		t.Snooze = snooze.String
		t.Triage = triage.String
		if updatedAt != "" {
			t.ModTime, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if deps, _ := d.GetTicketDeps(t.ID); deps != nil {
			t.Deps = deps
		}
		if tags, _ := d.GetTicketTags(t.ID); tags != nil {
			t.Tags = tags
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

// ListTicketsByDir returns all tickets belonging to the project registered
// for the given tickets directory (absolute path). Includes all statuses.
func (d *DB) ListTicketsByDir(ticketsDir string) ([]*Ticket, error) {
	q := `SELECT t.ticket_id, t.title, t.status, t.type, t.priority,
		         t.assignee, parent.ticket_id, t.external_ref, t.snooze, t.triage,
		         t.created_at, t.updated_at, t.body
		  FROM tickets t
		  JOIN projects p ON t.project_id = p.id
		  LEFT JOIN tickets parent ON t.parent_id = parent.id
		  WHERE p.tickets_dir = ?
		  ORDER BY t.priority, t.updated_at DESC`

	rows, err := d.db.Query(q, ticketsDir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		t := &Ticket{
			Deps: []string{},
			Tags: []string{},
		}
		var assignee, parentTicketID, extRef, snooze, triage sql.NullString
		var updatedAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.Type, &t.Priority,
			&assignee, &parentTicketID, &extRef, &snooze, &triage,
			&t.Created, &updatedAt, &t.Body); err != nil {
			return nil, err
		}
		t.Assignee = assignee.String
		t.Parent = parentTicketID.String
		t.ExternalRef = extRef.String
		t.Snooze = snooze.String
		t.Triage = triage.String
		if updatedAt != "" {
			t.ModTime, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if deps, _ := d.GetTicketDeps(t.ID); deps != nil {
			t.Deps = deps
		}
		if tags, _ := d.GetTicketTags(t.ID); tags != nil {
			t.Tags = tags
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

// ResolveIDDB finds a ticket by exact ID match, then by substring match.
// If ticketsDir is non-empty, the search is scoped to that project first,
// with a global fallback for cross-project partial IDs.
func (d *DB) ResolveIDDB(ticketsDir, partial string) (string, error) {
	var abs string
	if ticketsDir != "" {
		if a, err := filepath.Abs(ticketsDir); err == nil {
			abs = a
		} else {
			abs = ticketsDir
		}
	}

	// Exact match: project-scoped first, then global.
	if abs != "" {
		var id string
		err := d.db.QueryRow(`SELECT t.ticket_id FROM tickets t
			JOIN projects p ON t.project_id = p.id
			WHERE p.tickets_dir = ? AND t.ticket_id = ?`, abs, partial).Scan(&id)
		if err == nil {
			return id, nil
		}
	}
	var id string
	if err := d.db.QueryRow("SELECT ticket_id FROM tickets WHERE ticket_id = ?", partial).Scan(&id); err == nil {
		return id, nil
	}

	// Substring match: project-scoped first, then global.
	like := "%" + partial + "%"
	var matches []string
	if abs != "" {
		rows, err := d.db.Query(`SELECT t.ticket_id FROM tickets t
			JOIN projects p ON t.project_id = p.id
			WHERE p.tickets_dir = ? AND t.ticket_id LIKE ?`, abs, like)
		if err == nil {
			for rows.Next() {
				var m string
				if err := rows.Scan(&m); err == nil {
					matches = append(matches, m)
				}
			}
			rows.Close()
		}
	}
	if len(matches) == 0 {
		rows, err := d.db.Query("SELECT ticket_id FROM tickets WHERE ticket_id LIKE ?", like)
		if err == nil {
			for rows.Next() {
				var m string
				if err := rows.Scan(&m); err == nil {
					matches = append(matches, m)
				}
			}
			rows.Close()
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("ticket '%s' not found", partial)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous ID '%s' matches: %s", partial, strings.Join(matches, ", "))
	}
}

// ListReadyDB returns ready tickets using the ready_tickets view.
func (d *DB) ListReadyDB(project string, limit int) ([]*Ticket, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "1=1")

	if project != "" {
		conditions = append(conditions, "p.tag = ?")
		args = append(args, project)
	}

	q := `SELECT r.ticket_id, r.title, r.status, r.type, r.priority,
		         r.assignee, parent.ticket_id, r.external_ref, r.snooze, r.triage,
		         r.created_at, r.updated_at, r.body
		  FROM ready_tickets r
		  JOIN projects p ON r.project_id = p.id
		  LEFT JOIN tickets parent ON r.parent_id = parent.id
		  WHERE ` + strings.Join(conditions, " AND ") + `
		  ORDER BY r.priority, r.updated_at DESC`

	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		t := &Ticket{
			Deps: []string{},
			Tags: []string{},
		}
		var assignee, parentTicketID, extRef, snooze, triage sql.NullString
		var updatedAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.Type, &t.Priority,
			&assignee, &parentTicketID, &extRef, &snooze, &triage,
			&t.Created, &updatedAt, &t.Body); err != nil {
			return nil, err
		}
		t.Assignee = assignee.String
		t.Parent = parentTicketID.String
		t.ExternalRef = extRef.String
		t.Snooze = snooze.String
		t.Triage = triage.String
		if updatedAt != "" {
			t.ModTime, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if deps, _ := d.GetTicketDeps(t.ID); deps != nil {
			t.Deps = deps
		}
		if tags, _ := d.GetTicketTags(t.ID); tags != nil {
			t.Tags = tags
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

// GetTicketDB returns a single ticket by ID.
func (d *DB) GetTicketDB(ticketID string) (*Ticket, error) {
	q := `SELECT t.ticket_id, t.title, t.status, t.type, t.priority,
		         t.assignee, parent.ticket_id, t.external_ref, t.snooze, t.triage,
		         t.created_at, t.updated_at, t.body
		  FROM tickets t
		  LEFT JOIN tickets parent ON t.parent_id = parent.id
		  WHERE t.ticket_id = ?`

	t := &Ticket{
		Deps: []string{},
		Tags: []string{},
	}
	var assignee, parentTicketID, extRef, snooze, triage sql.NullString
	var updatedAt string
	err := d.db.QueryRow(q, ticketID).Scan(&t.ID, &t.Title, &t.Status, &t.Type, &t.Priority,
		&assignee, &parentTicketID, &extRef, &snooze, &triage,
		&t.Created, &updatedAt, &t.Body)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}
	if err != nil {
		return nil, err
	}

	t.Assignee = assignee.String
	t.Parent = parentTicketID.String
	t.ExternalRef = extRef.String
	t.Snooze = snooze.String
	t.Triage = triage.String
	if updatedAt != "" {
		t.ModTime, _ = time.Parse(time.RFC3339, updatedAt)
	}
	if deps, _ := d.GetTicketDeps(t.ID); deps != nil {
		t.Deps = deps
	}
	if tags, _ := d.GetTicketTags(t.ID); tags != nil {
		t.Tags = tags
	}
	t.PlanQuestions, _ = d.GetPlanQuestions(t.ID)

	return t, nil
}

// GetTicketDeps returns dependency IDs for a ticket.
func (d *DB) GetTicketDeps(ticketID string) ([]string, error) {
	q := `SELECT d.depends_on FROM ticket_deps d
		  JOIN tickets t ON d.ticket_id = t.id
		  WHERE t.ticket_id = ?`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, nil
}

// GetTicketTags returns tags for a ticket.
func (d *DB) GetTicketTags(ticketID string) ([]string, error) {
	q := `SELECT tt.tag FROM ticket_tags tt
		  JOIN tickets t ON tt.ticket_id = t.id
		  WHERE t.ticket_id = ?
		  ORDER BY tt.rowid`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetOpenDepsDB returns dep IDs that are not in closed/resolved status.
func (d *DB) GetOpenDepsDB(ticketID string) ([]string, error) {
	q := `SELECT d.depends_on FROM ticket_deps d
		  JOIN tickets t ON d.ticket_id = t.id
		  LEFT JOIN tickets dep ON dep.ticket_id = d.depends_on
		  WHERE t.ticket_id = ?
		    AND (dep.id IS NULL OR dep.status NOT IN ('closed', 'resolved'))`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, nil
}

// GetBlockingDB returns IDs of tickets that depend on this ticket.
func (d *DB) GetBlockingDB(ticketID string) ([]string, error) {
	q := `SELECT t2.ticket_id FROM ticket_deps d
		  JOIN tickets t ON d.ticket_id = t.id
		  JOIN tickets t2 ON t2.id = d.ticket_id
		  WHERE d.depends_on = ?`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocking []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		blocking = append(blocking, id)
	}
	return blocking, nil
}

// GetChildrenDB returns IDs of tickets whose parent is this ticket.
func (d *DB) GetChildrenDB(ticketID string) ([]string, error) {
	q := `SELECT t.ticket_id FROM tickets t
		  JOIN tickets parent ON t.parent_id = parent.id
		  WHERE parent.ticket_id = ?`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		children = append(children, id)
	}
	return children, nil
}

// GetPlanQuestions returns plan questions for a ticket.
func (d *DB) GetPlanQuestions(ticketID string) ([]PlanQuestion, error) {
	q := `SELECT pq.id, pq.question_id, pq.question, pq.context
		  FROM plan_questions pq
		  JOIN tickets t ON pq.ticket_id = t.id
		  WHERE t.ticket_id = ?
		  ORDER BY pq.sort_order`
	rows, err := d.db.Query(q, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []PlanQuestion
	for rows.Next() {
		var dbID int
		var q PlanQuestion
		var ctx sql.NullString
		if err := rows.Scan(&dbID, &q.ID, &q.Question, &ctx); err != nil {
			return nil, err
		}
		q.Context = ctx.String
		q.Options, _ = d.getPlanQuestionOptions(dbID)
		questions = append(questions, q)
	}
	return questions, nil
}

func (d *DB) getPlanQuestionOptions(questionID int) ([]QuestionOption, error) {
	q := `SELECT label, value, description FROM plan_question_options
		  WHERE plan_question_id = ? ORDER BY sort_order`
	rows, err := d.db.Query(q, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []QuestionOption
	for rows.Next() {
		var o QuestionOption
		var desc sql.NullString
		if err := rows.Scan(&o.Label, &o.Value, &desc); err != nil {
			return nil, err
		}
		o.Description = desc.String
		options = append(options, o)
	}
	return options, nil
}

// AllDepsResolvedDB checks if all dependencies are resolved using SQLite.
func (d *DB) AllDepsResolvedDB(ticketID string) bool {
	openDeps, err := d.GetOpenDepsDB(ticketID)
	if err != nil {
		return false
	}
	return len(openDeps) == 0
}
