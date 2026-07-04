package main

import (
	"os"
	"path/filepath"
	"strings"
)

// QQLMapping maps ko project tags onto Questbook realm/campaign slugs. It is a
// static file (the shim never guesses); its shape is the documented contract
// shared with the Questbook bulk-import dispatch. See QQL_MAPPING.md.
//
// File format (YAML), default location ~/.config/knockout/qql-mapping.yaml,
// overridable via KO_QQL_MAPPING:
//
//	default_realm: knockout
//	projects:
//	  fort-nix:
//	    realm: fort-nix
//	    campaign: fort-nix-maintenance
//	  questbook:
//	    realm: questbook
//	    campaign: questbook-buildout
type QQLMapping struct {
	DefaultRealm string
	Projects     map[string]QQLProjectMap
}

// QQLProjectMap is one project's realm/campaign binding. Realm is the anchor
// (quests attach to it); Campaign is optional.
type QQLProjectMap struct {
	Realm    string
	Campaign string
}

// QQLMappingPath returns the mapping file path, honoring KO_QQL_MAPPING then
// XDG_CONFIG_HOME, defaulting to ~/.config/knockout/qql-mapping.yaml.
func QQLMappingPath() string {
	if p := os.Getenv("KO_QQL_MAPPING"); p != "" {
		return p
	}
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "knockout", "qql-mapping.yaml")
}

// LoadQQLMapping reads the mapping file. A missing file is not an error: it
// returns an empty mapping (the shim then falls back to the project tag as the
// realm slug), so the shim works out of the box during early cutover.
func LoadQQLMapping() (*QQLMapping, error) {
	path := QQLMappingPath()
	if path == "" {
		return &QQLMapping{Projects: map[string]QQLProjectMap{}}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &QQLMapping{Projects: map[string]QQLProjectMap{}}, nil
		}
		return nil, err
	}
	return ParseQQLMapping(string(data)), nil
}

// ParseQQLMapping parses the minimal YAML shape above. Kept deliberately small
// (no external YAML dep) — the file is flat and hand-maintained.
func ParseQQLMapping(content string) *QQLMapping {
	m := &QQLMapping{Projects: map[string]QQLProjectMap{}}
	var inProjects bool
	var current string

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if indent == 0 {
			key, val, ok := parseYAMLLine(trimmed)
			if !ok {
				continue
			}
			switch key {
			case "default_realm":
				m.DefaultRealm = unquote(val)
				inProjects = false
			case "projects":
				inProjects = true
			default:
				inProjects = false
			}
			current = ""
			continue
		}

		if !inProjects {
			continue
		}

		key, val, ok := parseYAMLLine(trimmed)
		if !ok {
			continue
		}
		if indent <= 2 {
			// Project tag line: "fort-nix:"
			current = key
			pm := m.Projects[current]
			if val != "" {
				// Inline "tag: realm" shorthand.
				pm.Realm = unquote(val)
			}
			m.Projects[current] = pm
			continue
		}
		// Nested realm/campaign under the current project.
		if current == "" {
			continue
		}
		pm := m.Projects[current]
		switch key {
		case "realm":
			pm.Realm = unquote(val)
		case "campaign":
			pm.Campaign = unquote(val)
		}
		m.Projects[current] = pm
	}
	return m
}

// Resolve returns the realm and campaign slugs for a project tag. If the
// project is unmapped, the realm defaults to default_realm, then to the tag
// itself — never empty, so a quest always has a realm anchor.
func (m *QQLMapping) Resolve(tag string) (realm, campaign string) {
	if pm, ok := m.Projects[tag]; ok {
		realm, campaign = pm.Realm, pm.Campaign
	}
	if realm == "" {
		realm = m.DefaultRealm
	}
	if realm == "" {
		realm = tag
	}
	return realm, campaign
}
