package main

import (
	"os"
	"path/filepath"
	"strings"
)

// GlobalConfig represents ~/.config/knockout/config.yaml — user-level settings
// that apply across all projects unless overridden at the project level.
type GlobalConfig struct {
	Summarizer string // command to summarize long titles (e.g., "ollama run qwen3:0.6b --nowordwrap")
}

// GlobalConfigPath returns the path to the global config file.
// Respects XDG_CONFIG_HOME, defaults to ~/.config.
func GlobalConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "knockout", "config.yaml")
}

// LoadGlobalConfig reads the global config from disk.
// Returns a zero-value config (not an error) if the file does not exist.
func LoadGlobalConfig() (*GlobalConfig, error) {
	path := GlobalConfigPath()
	if path == "" {
		return &GlobalConfig{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, err
	}
	return ParseGlobalConfig(string(data))
}

// ParseGlobalConfig parses global config YAML content.
func ParseGlobalConfig(content string) (*GlobalConfig, error) {
	g := &GlobalConfig{}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, val, ok := parseYAMLLine(trimmed)
		if !ok {
			continue
		}
		// Strip inline comments
		if idx := strings.Index(val, " #"); idx >= 0 {
			val = strings.TrimSpace(val[:idx])
		}
		switch key {
		case "summarizer":
			g.Summarizer = val
		}
	}
	return g, nil
}

// ResolveSummarizer returns the summarizer command to use for a given project.
// Project-level config overrides global. Returns "" if none configured.
func ResolveSummarizer(ticketsDir string) string {
	// Check project-level config first
	configPath, err := FindConfig(ticketsDir)
	if err == nil {
		config, err := LoadConfig(configPath)
		if err == nil && config.Summarizer != "" {
			return config.Summarizer
		}
	}

	// Fall back to global config
	global, err := LoadGlobalConfig()
	if err != nil || global.Summarizer == "" {
		return ""
	}
	return global.Summarizer
}
