package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Registry represents the project registry at ~/.config/knockout/projects.yml.
type Registry struct {
	Default  string
	Projects map[string]string // tag -> absolute path
	Prefixes map[string]string // tag -> ticket prefix (e.g. "fn" for fort-nix)
}

// RegistryPath returns the path to the registry file.
// Respects XDG_CONFIG_HOME, defaults to ~/.config.
func RegistryPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "knockout", "projects.yml")
}

// LoadRegistry reads the registry from disk.
// Returns an empty registry (not an error) if the file does not exist.
func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Projects: map[string]string{}, Prefixes: map[string]string{}}, nil
		}
		return nil, err
	}
	reg, err := ParseRegistry(string(data))
	if err != nil {
		return nil, err
	}
	// Lazy backfill: detect prefixes for projects that have tickets but no prefix
	if backfillPrefixes(reg) {
		SaveRegistry(path, reg)
	}
	return reg, nil
}

// ParseRegistry parses a registry from its YAML content.
func ParseRegistry(content string) (*Registry, error) {
	r := &Registry{Projects: map[string]string{}, Prefixes: map[string]string{}}
	var section string // "", "projects", "prefixes"

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "projects:" {
			section = "projects"
			continue
		}
		if trimmed == "prefixes:" {
			section = "prefixes"
			continue
		}

		key, val, ok := parseYAMLLine(trimmed)
		if !ok {
			continue
		}

		indented := strings.HasPrefix(line, "  ")

		if !indented {
			section = ""
			if key == "default" {
				r.Default = val
			}
			continue
		}

		switch section {
		case "projects":
			r.Projects[key] = val
		case "prefixes":
			r.Prefixes[key] = val
		}
	}

	return r, nil
}

// FormatRegistry serializes a registry to YAML.
func FormatRegistry(r *Registry) string {
	var b strings.Builder
	if r.Default != "" {
		b.WriteString(fmt.Sprintf("default: %s\n", r.Default))
	}
	b.WriteString("projects:\n")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(r.Projects))
	for k := range r.Projects {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s: %s\n", k, r.Projects[k]))
	}

	if len(r.Prefixes) > 0 {
		b.WriteString("prefixes:\n")
		pkeys := make([]string, 0, len(r.Prefixes))
		for k := range r.Prefixes {
			pkeys = append(pkeys, k)
		}
		sortStrings(pkeys)
		for _, k := range pkeys {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, r.Prefixes[k]))
		}
	}
	return b.String()
}

// SaveRegistry writes the registry to disk, creating directories as needed.
func SaveRegistry(path string, r *Registry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(FormatRegistry(r)), 0644)
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// CleanTag strips a leading '#' from a tag string.
func CleanTag(tag string) string {
	return strings.TrimPrefix(tag, "#")
}

// backfillPrefixes detects prefixes for projects that have tickets but no
// prefix entry. Returns true if any prefixes were added.
func backfillPrefixes(reg *Registry) bool {
	changed := false
	for tag, path := range reg.Projects {
		if _, ok := reg.Prefixes[tag]; ok {
			continue
		}
		ticketsDir := resolveTicketsDir(path)
		prefix := detectPrefixFromDir(ticketsDir)
		if prefix != "" {
			reg.Prefixes[tag] = prefix
			changed = true
		}
	}
	return changed
}

// detectPrefixFromDir checks .ko/prefix first, then scans a tickets directory
// for existing ticket files and extracts the prefix from the first root-level
// ticket ID found. Returns "" if no prefix can be determined.
func detectPrefixFromDir(ticketsDir string) string {
	if p := ReadPrefix(ticketsDir); p != "" {
		return p
	}
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		// Root ticket: prefix-hash (no dots)
		if !strings.Contains(id, ".") {
			if idx := strings.Index(id, "-"); idx > 0 {
				return id[:idx]
			}
		}
	}
	return ""
}

// resolveTicketsDir returns the tickets directory for a project root.
func resolveTicketsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".ko", "tickets")
}

// findProjectRoot returns the absolute path to the project root.
// If a tickets directory is found, derives root via ProjectRoot.
// Otherwise, returns the current working directory.
func findProjectRoot() (string, error) {
	ticketsDir, err := FindTicketsDir()
	if err == nil {
		return ProjectRoot(ticketsDir), nil
	}
	return os.Getwd()
}

// CrossProjectLookup returns a dep lookup function that checks the local
// tickets directory first, then falls back to searching registered projects.
// When prefixes are available, uses prefix-based routing for O(1) lookups.
func CrossProjectLookup(localTicketsDir string, reg *Registry) func(string) (string, bool) {
	// Build reverse index: prefix -> tickets dir
	prefixIndex := make(map[string]string)
	for tag, prefix := range reg.Prefixes {
		if path, ok := reg.Projects[tag]; ok {
			prefixIndex[prefix] = resolveTicketsDir(path)
		}
	}

	return func(id string) (string, bool) {
		// Try local first
		t, err := LoadTicket(localTicketsDir, id)
		if err == nil {
			return t.Status, true
		}

		// Extract prefix from dep ID and try direct lookup
		if prefix := extractPrefix(id); prefix != "" {
			if dir, ok := prefixIndex[prefix]; ok {
				t, err := LoadTicket(dir, id)
				if err == nil {
					return t.Status, true
				}
			}
		}

		// Fallback: scan all registered projects
		for _, path := range reg.Projects {
			remoteDir := resolveTicketsDir(path)
			t, err := LoadTicket(remoteDir, id)
			if err == nil {
				return t.Status, true
			}
		}
		return "", false
	}
}

// extractPrefix returns the prefix from a ticket ID (e.g. "fn" from "fn-a001").
// Returns "" if no prefix can be extracted.
func extractPrefix(id string) string {
	// For hierarchical IDs like "fn-a001.b002", the prefix is before the first hyphen
	// of the root segment (before any dots).
	root := id
	if dot := strings.Index(id, "."); dot >= 0 {
		root = id[:dot]
	}
	if idx := strings.Index(root, "-"); idx > 0 {
		return root[:idx]
	}
	return ""
}

// RoutingDecision describes where a ticket should be created.
type RoutingDecision struct {
	Title         string
	RoutingTag    string   // first #tag (cleaned), empty if none
	ExtraTags     []string // remaining #tags (cleaned)
	TargetPath    string   // project path to create ticket in
	Status        string   // status for the created ticket
	IsRouted      bool     // true if routed to a different project
	IsCaptured    bool     // true if tag was unrecognized (sent to default)
}

// ParseTags extracts #tags from a title string.
// Returns the cleaned title and the list of tags (without #).
// Words starting with \# are treated as literal hashtags (escape is stripped).
func ParseTags(title string) (string, []string) {
	words := strings.Fields(title)
	var clean []string
	var tags []string
	for _, w := range words {
		if strings.HasPrefix(w, "\\#") {
			// Escaped hashtag — literal, strip the backslash
			clean = append(clean, w[1:])
		} else if strings.HasPrefix(w, "#") && len(w) > 1 {
			tags = append(tags, CleanTag(w))
		} else {
			clean = append(clean, w)
		}
	}
	return strings.Join(clean, " "), tags
}

// RouteTicket determines where a ticket should go based on its tags and
// the project registry. Pure decision function — no I/O.
func RouteTicket(title string, reg *Registry, localPath string) RoutingDecision {
	cleanTitle, tags := ParseTags(title)

	// No tags — local ticket
	if len(tags) == 0 {
		return RoutingDecision{
			Title:      cleanTitle,
			TargetPath: localPath,
			Status:     "open",
		}
	}

	routingTag := tags[0]
	extraTags := tags[1:]

	// Recognized tag — route to that project
	if path, ok := reg.Projects[routingTag]; ok {
		return RoutingDecision{
			Title:      cleanTitle,
			RoutingTag: routingTag,
			ExtraTags:  extraTags,
			TargetPath: path,
			Status:     "routed",
			IsRouted:   true,
		}
	}

	// Unrecognized tag — send to default project as captured
	if reg.Default != "" {
		if path, ok := reg.Projects[reg.Default]; ok {
			return RoutingDecision{
				Title:      cleanTitle,
				RoutingTag: routingTag,
				ExtraTags:  extraTags,
				TargetPath: path,
				Status:     "captured",
				IsCaptured: true,
			}
		}
	}

	// No default — create locally with tag preserved
	return RoutingDecision{
		Title:      cleanTitle,
		RoutingTag: routingTag,
		ExtraTags:  extraTags,
		TargetPath: localPath,
		Status:     "open",
	}
}
