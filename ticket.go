package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Statuses is the closed set of valid ticket statuses.
var Statuses = []string{"captured", "routed", "open", "in_progress", "closed", "blocked"}

// Ticket represents a ticket parsed from a markdown file with YAML frontmatter.
type Ticket struct {
	ID          string   `yaml:"id"`
	Status      string   `yaml:"status"`
	Deps        []string `yaml:"deps"`
	Links       []string `yaml:"links"`
	Created     string   `yaml:"created"`
	Type        string   `yaml:"type"`
	Priority    int      `yaml:"priority"`
	Assignee    string   `yaml:"assignee,omitempty"`
	Parent      string   `yaml:"parent,omitempty"`
	ExternalRef string   `yaml:"external-ref,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`

	// Title is extracted from the first markdown heading.
	Title string `yaml:"-"`
	// Body is everything after the frontmatter and title.
	Body string `yaml:"-"`
}

// ValidStatus reports whether s is a valid ticket status.
func ValidStatus(s string) bool {
	for _, v := range Statuses {
		if s == v {
			return true
		}
	}
	return false
}

// Depth returns the decomposition depth of a ticket ID (number of dots).
func Depth(id string) int {
	return strings.Count(id, ".")
}

// GenerateHash returns a short random hex string for ticket IDs.
func GenerateHash() string {
	b := make([]byte, 2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// IsReady reports whether a ticket should appear in the ready queue.
// Pure decision function: takes ticket status and whether all deps are resolved.
func IsReady(status string, allDepsResolved bool) bool {
	switch status {
	case "open", "in_progress":
		return allDepsResolved
	default:
		return false
	}
}

// FormatTicket serializes a Ticket to its markdown file representation.
func FormatTicket(t *Ticket) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", t.ID))
	b.WriteString(fmt.Sprintf("status: %s\n", t.Status))
	b.WriteString(fmt.Sprintf("deps: [%s]\n", strings.Join(t.Deps, ", ")))
	b.WriteString(fmt.Sprintf("links: [%s]\n", strings.Join(t.Links, ", ")))
	b.WriteString(fmt.Sprintf("created: %s\n", t.Created))
	b.WriteString(fmt.Sprintf("type: %s\n", t.Type))
	b.WriteString(fmt.Sprintf("priority: %d\n", t.Priority))
	if t.Assignee != "" {
		b.WriteString(fmt.Sprintf("assignee: %s\n", t.Assignee))
	}
	if t.Parent != "" {
		b.WriteString(fmt.Sprintf("parent: %s\n", t.Parent))
	}
	if t.ExternalRef != "" {
		b.WriteString(fmt.Sprintf("external-ref: %s\n", t.ExternalRef))
	}
	if len(t.Tags) > 0 {
		b.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(t.Tags, ", ")))
	}
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("# %s\n", t.Title))
	if t.Body != "" {
		b.WriteString(t.Body)
	}
	return b.String()
}

// ParseTicket parses a ticket from its markdown file content.
func ParseTicket(content string) (*Ticket, error) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing frontmatter")
	}
	rest := content[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return nil, fmt.Errorf("unterminated frontmatter")
	}
	frontmatter := rest[:end]
	body := rest[end+5:] // skip \n---\n

	t := &Ticket{
		Deps:  []string{},
		Links: []string{},
	}

	for _, line := range strings.Split(frontmatter, "\n") {
		key, val, ok := parseYAMLLine(line)
		if !ok {
			continue
		}
		switch key {
		case "id":
			t.ID = val
		case "status":
			t.Status = val
		case "deps":
			t.Deps = parseYAMLList(val)
		case "links":
			t.Links = parseYAMLList(val)
		case "created":
			t.Created = val
		case "type":
			t.Type = val
		case "priority":
			fmt.Sscanf(val, "%d", &t.Priority)
		case "assignee":
			t.Assignee = val
		case "parent":
			t.Parent = val
		case "external-ref":
			t.ExternalRef = val
		case "tags":
			t.Tags = parseYAMLList(val)
		}
	}

	// Extract title from first heading
	lines := strings.SplitN(body, "\n", 2)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		t.Title = strings.TrimPrefix(lines[0], "# ")
		if len(lines) > 1 {
			t.Body = lines[1]
		}
	} else {
		t.Body = body
	}

	return t, nil
}

// parseYAMLLine does minimal YAML parsing for "key: value" lines.
func parseYAMLLine(line string) (key, val string, ok bool) {
	idx := strings.Index(line, ": ")
	if idx < 0 {
		// Handle "key:" with no value
		if strings.HasSuffix(line, ":") {
			return strings.TrimSpace(line[:len(line)-1]), "", true
		}
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+2:]), true
}

// parseYAMLList parses "[a, b, c]" into a string slice. Returns nil for "[]".
func parseYAMLList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "[]" || s == "" {
		return []string{}
	}
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// TicketPath returns the file path for a ticket given the tickets directory.
func TicketPath(ticketsDir, id string) string {
	return filepath.Join(ticketsDir, id+".md")
}

// LoadTicket reads and parses a ticket from disk.
func LoadTicket(ticketsDir, id string) (*Ticket, error) {
	data, err := os.ReadFile(TicketPath(ticketsDir, id))
	if err != nil {
		return nil, fmt.Errorf("ticket '%s' not found", id)
	}
	return ParseTicket(string(data))
}

// SaveTicket writes a ticket to disk.
func SaveTicket(ticketsDir string, t *Ticket) error {
	return os.WriteFile(TicketPath(ticketsDir, t.ID), []byte(FormatTicket(t)), 0644)
}

// ListTickets reads all tickets from the tickets directory.
func ListTickets(ticketsDir string) ([]*Ticket, error) {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tickets []*Ticket
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ticketsDir, e.Name()))
		if err != nil {
			continue
		}
		t, err := ParseTicket(string(data))
		if err != nil {
			continue
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

// ResolveID finds a ticket by exact match, then by substring match.
// Returns an error if no match or ambiguous.
func ResolveID(ticketsDir, partial string) (string, error) {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return "", fmt.Errorf("ticket '%s' not found", partial)
	}

	var ids []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			ids = append(ids, strings.TrimSuffix(e.Name(), ".md"))
		}
	}

	// Exact match first
	for _, id := range ids {
		if id == partial {
			return id, nil
		}
	}

	// Substring match
	var matches []string
	for _, id := range ids {
		if strings.Contains(id, partial) {
			matches = append(matches, id)
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

// FindTicketsDir walks up from the current directory looking for the tickets
// directory. Checks .ko/tickets/ first (new layout), then .tickets/ (legacy).
// If only the legacy .tickets/ exists, it is migrated to .ko/tickets/.
// Respects TICKETS_DIR env var.
func FindTicketsDir() (string, error) {
	if env := os.Getenv("TICKETS_DIR"); env != "" {
		info, err := os.Stat(env)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("TICKETS_DIR '%s' is not a directory", env)
		}
		return env, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		// Prefer new layout
		newPath := filepath.Join(dir, ".ko", "tickets")
		if info, err := os.Stat(newPath); err == nil && info.IsDir() {
			return newPath, nil
		}
		// Check legacy layout and migrate
		oldPath := filepath.Join(dir, ".tickets")
		if info, err := os.Stat(oldPath); err == nil && info.IsDir() {
			if migrated, err := migrateTicketsDir(dir); err == nil {
				return migrated, nil
			}
			// Migration failed — use legacy path
			return oldPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .ko/tickets directory found")
}

// migrateTicketsDir moves .tickets/ to .ko/tickets/ and returns the new path.
func migrateTicketsDir(projectRoot string) (string, error) {
	oldPath := filepath.Join(projectRoot, ".tickets")
	newPath := filepath.Join(projectRoot, ".ko", "tickets")

	// Ensure .ko/ exists
	if err := os.MkdirAll(filepath.Join(projectRoot, ".ko"), 0755); err != nil {
		return "", err
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return "", err
	}

	fmt.Fprintf(os.Stderr, "ko: migrated .tickets/ -> .ko/tickets/\n")
	return newPath, nil
}

// ProjectRoot returns the project root directory for a given tickets directory.
// Handles both .ko/tickets/ (new) and .tickets/ (legacy) layouts.
func ProjectRoot(ticketsDir string) string {
	abs, err := filepath.Abs(ticketsDir)
	if err != nil {
		return filepath.Dir(ticketsDir)
	}
	// .ko/tickets/ -> go up two levels
	if filepath.Base(filepath.Dir(abs)) == ".ko" {
		return filepath.Dir(filepath.Dir(abs))
	}
	// .tickets/ -> go up one level
	return filepath.Dir(abs)
}

// EnsureTicketsDir creates the .tickets directory if it doesn't exist.
func EnsureTicketsDir(ticketsDir string) error {
	return os.MkdirAll(ticketsDir, 0755)
}

// NewTicket creates a Ticket with defaults.
func NewTicket(prefix, title string) *Ticket {
	hash := GenerateHash()
	id := prefix + "-" + hash
	return &Ticket{
		ID:       id,
		Title:    title,
		Status:   "open",
		Deps:     []string{},
		Links:    []string{},
		Created:  time.Now().UTC().Format(time.RFC3339),
		Type:     "task",
		Priority: 2,
	}
}

// NewChildTicket creates a child ticket under a parent ID.
func NewChildTicket(parentID, title string) *Ticket {
	hash := GenerateHash()
	id := parentID + "." + hash
	t := &Ticket{
		ID:       id,
		Title:    title,
		Status:   "open",
		Deps:     []string{},
		Links:    []string{},
		Created:  time.Now().UTC().Format(time.RFC3339),
		Type:     "task",
		Priority: 2,
		Parent:   parentID,
	}
	return t
}

// SortByPriorityThenID sorts tickets by priority (ascending) then ID (ascending).
func SortByPriorityThenID(tickets []*Ticket) {
	sort.Slice(tickets, func(i, j int) bool {
		if tickets[i].Priority != tickets[j].Priority {
			return tickets[i].Priority < tickets[j].Priority
		}
		return tickets[i].ID < tickets[j].ID
	})
}

// AllDepsResolved checks if all deps of a ticket are closed.
func AllDepsResolved(ticketsDir string, deps []string) bool {
	return AllDepsResolvedWith(deps, func(id string) (string, bool) {
		t, err := LoadTicket(ticketsDir, id)
		if err != nil {
			return "", false
		}
		return t.Status, true
	})
}

// AllDepsResolvedWith checks if all deps are closed using a lookup function.
// The lookup returns (status, found). Pure decision function.
func AllDepsResolvedWith(deps []string, lookup func(id string) (string, bool)) bool {
	for _, depID := range deps {
		status, found := lookup(depID)
		if !found || status != "closed" {
			return false
		}
	}
	return true
}

// AddNote appends a timestamped note to the ticket body.
func AddNote(t *Ticket, note string) {
	ts := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	section := fmt.Sprintf("\n## Notes\n\n**%s:** %s\n", ts, note)
	if strings.Contains(t.Body, "## Notes") {
		// Append to existing notes section
		t.Body += fmt.Sprintf("\n**%s:** %s\n", ts, note)
	} else {
		t.Body += section
	}
}
