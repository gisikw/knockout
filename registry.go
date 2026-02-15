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
