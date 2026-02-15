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
			return &Registry{Projects: map[string]string{}}, nil
		}
		return nil, err
	}
	return ParseRegistry(string(data))
}

// ParseRegistry parses a registry from its YAML content.
func ParseRegistry(content string) (*Registry, error) {
	r := &Registry{Projects: map[string]string{}}
	inProjects := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "projects:" {
			inProjects = true
			continue
		}

		key, val, ok := parseYAMLLine(trimmed)
		if !ok {
			continue
		}

		if !inProjects {
			if key == "default" {
				r.Default = val
			}
			continue
		}

		// Inside projects: indented "tag: path" lines
		if strings.HasPrefix(line, "  ") {
			r.Projects[key] = val
		} else {
			// No longer indented — left the projects block
			inProjects = false
			if key == "default" {
				r.Default = val
			}
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

// CrossProjectLookup returns a dep lookup function that checks the local
// tickets directory first, then falls back to searching registered projects.
func CrossProjectLookup(localTicketsDir string, reg *Registry) func(string) (string, bool) {
	return func(id string) (string, bool) {
		// Try local first
		t, err := LoadTicket(localTicketsDir, id)
		if err == nil {
			return t.Status, true
		}
		// Search registered projects
		for _, path := range reg.Projects {
			remoteDir := filepath.Join(path, ".tickets")
			t, err := LoadTicket(remoteDir, id)
			if err == nil {
				return t.Status, true
			}
		}
		return "", false
	}
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
func ParseTags(title string) (string, []string) {
	words := strings.Fields(title)
	var clean []string
	var tags []string
	for _, w := range words {
		if strings.HasPrefix(w, "#") && len(w) > 1 {
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
