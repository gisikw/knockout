package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrNoLocalProject is returned by FindTicketsDir when no .ko/tickets directory
// is found by walking up from cwd. This is a soft error — callers that can
// resolve tickets by prefix should treat it as non-fatal.
var ErrNoLocalProject = errors.New("no .ko/tickets directory found")

// Statuses is the closed set of valid ticket statuses.
var Statuses = []string{"captured", "routed", "open", "in_progress", "closed", "blocked", "resolved"}

// Ticket represents a ticket parsed from a markdown file with YAML frontmatter.
type Ticket struct {
	ID            string         `yaml:"id"`
	Status        string         `yaml:"status"`
	Deps          []string       `yaml:"deps"`
	Created       string         `yaml:"created"`
	Type          string         `yaml:"type"`
	Priority      int            `yaml:"priority"`
	Assignee      string         `yaml:"assignee,omitempty"`
	Parent        string         `yaml:"parent,omitempty"`
	ExternalRef   string         `yaml:"external-ref,omitempty"`
	Snooze        string         `yaml:"snooze,omitempty"`
	Tags          []string       `yaml:"tags,omitempty"`
	PlanQuestions []PlanQuestion `yaml:"plan-questions,omitempty"`

	// Title is extracted from the first markdown heading.
	Title string `yaml:"-"`
	// Body is everything after the frontmatter and title.
	Body string `yaml:"-"`
	// ModTime is the file modification time, populated by ListTickets/LoadTicket.
	ModTime time.Time `yaml:"-"`
}

// PlanQuestion represents a question that needs to be answered before implementing a ticket.
type PlanQuestion struct {
	ID       string           `yaml:"id" json:"id"`
	Question string           `yaml:"question" json:"question"`
	Context  string           `yaml:"context,omitempty" json:"context,omitempty"`
	Options  []QuestionOption `yaml:"options" json:"options"`
}

// QuestionOption represents a possible answer to a plan question.
type QuestionOption struct {
	Label       string `yaml:"label" json:"label"`
	Value       string `yaml:"value" json:"value"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
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
	case "resolved", "closed":
		return false
	default:
		return false
	}
}

// IsSnoozed reports whether a ticket is currently snoozed.
// A ticket is snoozed when snooze is non-empty and now is before midnight UTC
// on the parsed snooze date. Empty or unparseable snooze strings return false.
func IsSnoozed(snooze string, now time.Time) bool {
	if snooze == "" {
		return false
	}
	t, err := time.ParseInLocation("2006-01-02", snooze, time.UTC)
	if err != nil {
		return false
	}
	return now.Before(t)
}

// FormatTicket serializes a Ticket to its markdown file representation.
func FormatTicket(t *Ticket) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", t.ID))
	b.WriteString(fmt.Sprintf("status: %s\n", t.Status))
	b.WriteString(fmt.Sprintf("deps: [%s]\n", strings.Join(t.Deps, ", ")))
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
	if t.Snooze != "" {
		b.WriteString(fmt.Sprintf("snooze: %s\n", t.Snooze))
	}
	if len(t.Tags) > 0 {
		b.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(t.Tags, ", ")))
	}
	if len(t.PlanQuestions) > 0 {
		b.WriteString("plan-questions:\n")
		for _, q := range t.PlanQuestions {
			b.WriteString(fmt.Sprintf("  - id: %s\n", q.ID))
			b.WriteString(fmt.Sprintf("    question: \"%s\"\n", q.Question))
			if q.Context != "" {
				b.WriteString(fmt.Sprintf("    context: \"%s\"\n", q.Context))
			}
			b.WriteString("    options:\n")
			for _, opt := range q.Options {
				b.WriteString(fmt.Sprintf("      - label: \"%s\"\n", opt.Label))
				b.WriteString(fmt.Sprintf("        value: %s\n", opt.Value))
				if opt.Description != "" {
					b.WriteString(fmt.Sprintf("        description: \"%s\"\n", opt.Description))
				}
			}
		}
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
		Deps: []string{},
	}

	lines := strings.Split(frontmatter, "\n")
	var inPlanQuestions bool
	var currentQuestion *PlanQuestion
	var currentOption *QuestionOption
	var inOptions bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		indent := countIndent(line)

		// Detect plan-questions section
		if indent == 0 && trimmed == "plan-questions:" {
			inPlanQuestions = true
			continue
		}

		// Exit plan-questions when we hit another top-level key
		if indent == 0 && inPlanQuestions && !strings.HasPrefix(trimmed, "-") {
			inPlanQuestions = false
			currentQuestion = nil
			currentOption = nil
			inOptions = false
		}

		if inPlanQuestions {
			switch {
			case indent == 2 && strings.HasPrefix(trimmed, "- "):
				// New question item
				if currentQuestion != nil {
					if currentOption != nil {
						currentQuestion.Options = append(currentQuestion.Options, *currentOption)
						currentOption = nil
					}
					t.PlanQuestions = append(t.PlanQuestions, *currentQuestion)
				}
				currentQuestion = &PlanQuestion{}
				inOptions = false
				// Parse inline key-value if present (e.g., "- id: q1")
				rest := strings.TrimPrefix(trimmed, "- ")
				if key, val, ok := parseYAMLLine(rest); ok {
					switch key {
					case "id":
						currentQuestion.ID = unquote(val)
					case "question":
						currentQuestion.Question = unquote(val)
					case "context":
						currentQuestion.Context = unquote(val)
					}
				}

			case indent == 4 && currentQuestion != nil && !inOptions:
				// Question properties
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				switch key {
				case "id":
					currentQuestion.ID = unquote(val)
				case "question":
					currentQuestion.Question = unquote(val)
				case "context":
					currentQuestion.Context = unquote(val)
				case "options":
					inOptions = true
				}

			case indent == 6 && inOptions && strings.HasPrefix(trimmed, "- "):
				// New option item
				if currentOption != nil {
					currentQuestion.Options = append(currentQuestion.Options, *currentOption)
				}
				currentOption = &QuestionOption{}
				// Parse inline key-value if present
				rest := strings.TrimPrefix(trimmed, "- ")
				if key, val, ok := parseYAMLLine(rest); ok {
					switch key {
					case "label":
						currentOption.Label = unquote(val)
					case "value":
						currentOption.Value = unquote(val)
					case "description":
						currentOption.Description = unquote(val)
					}
				}

			case indent == 8 && currentOption != nil:
				// Option properties
				key, val, ok := parseYAMLLine(trimmed)
				if !ok {
					continue
				}
				switch key {
				case "label":
					currentOption.Label = unquote(val)
				case "value":
					currentOption.Value = unquote(val)
				case "description":
					currentOption.Description = unquote(val)
				}
			}
			continue
		}

		// Standard frontmatter parsing
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
		case "snooze":
			t.Snooze = val
		case "tags":
			t.Tags = parseYAMLList(val)
		}
	}

	// Flush any pending question/option
	if currentOption != nil && currentQuestion != nil {
		currentQuestion.Options = append(currentQuestion.Options, *currentOption)
	}
	if currentQuestion != nil {
		t.PlanQuestions = append(t.PlanQuestions, *currentQuestion)
	}

	// Extract title from first heading
	bodyLines := strings.SplitN(body, "\n", 2)
	if len(bodyLines) > 0 && strings.HasPrefix(bodyLines[0], "# ") {
		t.Title = strings.TrimPrefix(bodyLines[0], "# ")
		if len(bodyLines) > 1 {
			t.Body = bodyLines[1]
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

// unquote removes surrounding quotes from a string if present.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// TicketPath returns the file path for a ticket given the tickets directory.
func TicketPath(ticketsDir, id string) string {
	return filepath.Join(ticketsDir, id+".md")
}

// ArtifactDir returns the artifact directory path for a ticket.
func ArtifactDir(ticketsDir, id string) string {
	return filepath.Join(ticketsDir, id+".artifacts")
}

// EnsureArtifactDir creates the artifact directory for a ticket if it doesn't exist.
func EnsureArtifactDir(ticketsDir, id string) (string, error) {
	dir := ArtifactDir(ticketsDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// RemoveArtifactDir removes the artifact directory for a ticket.
func RemoveArtifactDir(ticketsDir, id string) {
	os.RemoveAll(ArtifactDir(ticketsDir, id))
}

// LoadTicket reads and parses a ticket from disk.
func LoadTicket(ticketsDir, id string) (*Ticket, error) {
	path := TicketPath(ticketsDir, id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ticket '%s' not found", id)
	}
	t, err := ParseTicket(string(data))
	if err != nil {
		return nil, err
	}
	if info, statErr := os.Stat(path); statErr == nil {
		t.ModTime = info.ModTime()
	}
	return t, nil
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
		path := filepath.Join(ticketsDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		t, err := ParseTicket(string(data))
		if err != nil {
			continue
		}
		if info, statErr := e.Info(); statErr == nil {
			t.ModTime = info.ModTime()
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

// ResolveTicket tries to resolve a partial ticket ID, first locally in ticketsDir,
// then by prefix-based lookup across the project registry. Returns the tickets
// directory where the ticket was found, the resolved full ID, and any error.
func ResolveTicket(ticketsDir, partial string) (string, string, error) {
	// Try local first
	id, err := ResolveID(ticketsDir, partial)
	if err == nil {
		return ticketsDir, id, nil
	}
	localErr := err

	// Extract prefix from the ticket ID
	prefix := extractPrefix(partial)
	if prefix == "" {
		return "", "", localErr
	}

	// Load registry
	regPath := RegistryPath()
	if regPath == "" {
		return "", "", localErr
	}
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return "", "", localErr
	}

	// Build reverse index: prefix -> ticketsDir
	for tag, p := range reg.Prefixes {
		if p == prefix {
			if projectPath, ok := reg.Projects[tag]; ok {
				remoteDir := resolveTicketsDir(projectPath)
				if id, err := ResolveID(remoteDir, partial); err == nil {
					return remoteDir, id, nil
				}
			}
		}
	}

	return "", "", localErr
}

// FindTicketsDir walks up from the current directory looking for .ko/tickets/.
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
		candidate := filepath.Join(dir, ".ko", "tickets")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ErrNoLocalProject
}

// ProjectRoot returns the project root for a tickets directory (.ko/tickets/ -> two levels up).
func ProjectRoot(ticketsDir string) string {
	abs, err := filepath.Abs(ticketsDir)
	if err != nil {
		return filepath.Dir(ticketsDir)
	}
	return filepath.Dir(filepath.Dir(abs))
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

// SortByPriorityThenModified sorts tickets by priority (ascending),
// then status (open before blocked), then ModTime (descending, newest first).
func SortByPriorityThenModified(tickets []*Ticket) {
	sort.Slice(tickets, func(i, j int) bool {
		if tickets[i].Priority != tickets[j].Priority {
			return tickets[i].Priority < tickets[j].Priority
		}
		oi := statusOrder(tickets[i].Status)
		oj := statusOrder(tickets[j].Status)
		if oi != oj {
			return oi < oj
		}
		return tickets[i].ModTime.After(tickets[j].ModTime)
	})
}

// statusOrder returns a sort rank: in_progress < open < blocked/etc < resolved < closed.
// Tickets with status "open" but unresolved deps are still "open" here —
// they sort with other open tickets, not with explicitly blocked ones.
func statusOrder(status string) int {
	switch status {
	case "in_progress":
		return 0
	case "open":
		return 1
	case "resolved":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
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

// AllDepsResolvedWith checks if all deps are in a terminal state using a lookup
// function. The lookup returns (status, found). Both "closed" and "resolved"
// count as terminal.
func AllDepsResolvedWith(deps []string, lookup func(id string) (string, bool)) bool {
	for _, depID := range deps {
		status, found := lookup(depID)
		if !found || (status != "closed" && status != "resolved") {
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
