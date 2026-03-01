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

// isOldFormat returns true if content uses the old flat format (has a top-level
// "prefixes:" section or a top-level "default:" key).
func isOldFormat(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "prefixes:") || strings.HasPrefix(line, "default:") {
			return true
		}
	}
	return false
}

// LoadRegistry reads the registry from disk.
// Returns an empty registry (not an error) if the file does not exist.
// Auto-migrates old-format files to the new nested format on first read.
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
	// Lazy backfill: detect prefixes for projects that have tickets but no prefix.
	// Also auto-migrate old-format files to new nested format.
	if backfillPrefixes(reg) || isOldFormat(string(data)) {
		SaveRegistry(path, reg)
	}
	return reg, nil
}

// ParseRegistry parses a registry from its YAML content.
// Handles both the old flat format and the new nested format.
func ParseRegistry(content string) (*Registry, error) {
	r := &Registry{Projects: map[string]string{}, Prefixes: map[string]string{}}
	var section string        // "", "projects", "prefixes"
	var currentProject string // non-empty when inside a new-format nested project block

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "projects:" {
			section = "projects"
			currentProject = ""
			continue
		}
		if trimmed == "prefixes:" {
			section = "prefixes"
			currentProject = ""
			continue
		}

		key, val, ok := parseYAMLLine(trimmed)
		if !ok {
			continue
		}

		is4space := strings.HasPrefix(line, "    ")
		is2space := strings.HasPrefix(line, "  ") && !is4space

		if is4space {
			// Properties of a nested project block (new format)
			if section == "projects" && currentProject != "" {
				switch key {
				case "path":
					r.Projects[currentProject] = val
				case "prefix":
					r.Prefixes[currentProject] = val
				case "default":
					if val == "true" {
						r.Default = currentProject
					}
				}
			}
			continue
		}

		if is2space {
			switch section {
			case "projects":
				if val == "" {
					// New format: bare project tag, start nested block
					currentProject = key
				} else {
					// Old format: tag: path
					r.Projects[key] = val
					currentProject = ""
				}
			case "prefixes":
				r.Prefixes[key] = val
			}
			continue
		}

		// Top-level (no indent): reset section
		section = ""
		currentProject = ""
		if key == "default" {
			r.Default = val
		}
	}

	return r, nil
}

// FormatRegistry serializes a registry to YAML using the nested format.
func FormatRegistry(r *Registry) string {
	var b strings.Builder
	b.WriteString("projects:\n")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(r.Projects))
	for k := range r.Projects {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s:\n", k))
		b.WriteString(fmt.Sprintf("    path: %s\n", r.Projects[k]))
		if prefix := r.Prefixes[k]; prefix != "" {
			b.WriteString(fmt.Sprintf("    prefix: %s\n", prefix))
		}
		if r.Default == k {
			b.WriteString("    default: true\n")
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

