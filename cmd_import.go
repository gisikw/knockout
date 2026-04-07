package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// cmdImport backfills all existing tickets and build history into the shadow database.
// Usage: ko import [--project=tag]
func cmdImport(args []string) int {
	// Parse optional --project flag.
	var projectTag string
	for _, a := range args {
		if strings.HasPrefix(a, "--project=") {
			projectTag = strings.TrimPrefix(a, "--project=")
		}
	}

	db := getShadowDB()
	if db == nil {
		fmt.Fprintln(os.Stderr, "ko: cannot open shadow database")
		return 1
	}

	if projectTag != "" {
		return importProject(db, projectTag)
	}
	return importAll(db)
}

// importAll imports every registered project plus the local project.
func importAll(db *DB) int {
	regPath := RegistryPath()
	if regPath != "" {
		reg, err := LoadRegistry(regPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko: load registry: %v\n", err)
			return 1
		}
		// Sync registry to DB first.
		if err := db.SyncRegistry(reg); err != nil {
			fmt.Fprintf(os.Stderr, "ko: sync registry: %v\n", err)
			return 1
		}
		for tag, path := range reg.Projects {
			ticketsDir := resolveTicketsDir(path)
			fmt.Fprintf(os.Stderr, "importing %s (%s)...\n", tag, ticketsDir)
			importTicketsDir(db, ticketsDir)
		}
	}

	// Also import local project if not already covered.
	if ticketsDir, err := FindTicketsDir(); err == nil {
		fmt.Fprintf(os.Stderr, "importing local (%s)...\n", ticketsDir)
		importTicketsDir(db, ticketsDir)
	}

	fmt.Fprintln(os.Stderr, "import complete")
	return 0
}

// importProject imports a single project by tag.
func importProject(db *DB, tag string) int {
	regPath := RegistryPath()
	if regPath == "" {
		fmt.Fprintln(os.Stderr, "ko: no registry found")
		return 1
	}
	reg, err := LoadRegistry(regPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: load registry: %v\n", err)
		return 1
	}
	path, ok := reg.Projects[tag]
	if !ok {
		fmt.Fprintf(os.Stderr, "ko: project '%s' not found in registry\n", tag)
		return 1
	}

	if err := db.SyncRegistry(reg); err != nil {
		fmt.Fprintf(os.Stderr, "ko: sync registry: %v\n", err)
		return 1
	}

	ticketsDir := resolveTicketsDir(path)
	fmt.Fprintf(os.Stderr, "importing %s (%s)...\n", tag, ticketsDir)
	importTicketsDir(db, ticketsDir)
	fmt.Fprintln(os.Stderr, "import complete")
	return 0
}

// importTicketsDir imports all tickets and build history from a tickets directory.
func importTicketsDir(db *DB, ticketsDir string) {
	tickets, err := ListTickets(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error listing tickets: %v\n", err)
		return
	}

	// Sort by depth so parents are imported before children (FK on parent_id).
	sort.Slice(tickets, func(i, j int) bool {
		di, dj := Depth(tickets[i].ID), Depth(tickets[j].ID)
		if di != dj {
			return di < dj
		}
		return tickets[i].ID < tickets[j].ID
	})

	var imported, failed int
	for _, t := range tickets {
		if err := db.UpsertTicket(t, ticketsDir); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", t.ID, err)
			failed++
			continue
		}
		imported++

		// Import build history JSONL if it exists.
		jsonlPath := BuildHistoryPath(ticketsDir, t.ID)
		importBuildHistory(db, t.ID, jsonlPath)
	}

	fmt.Fprintf(os.Stderr, "  %d tickets imported, %d failed\n", imported, failed)
}

// importBuildHistory imports a per-ticket .jsonl build history file.
func importBuildHistory(db *DB, ticketID, jsonlPath string) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return // no build history is fine
	}
	defer f.Close()

	prefix := extractPrefix(ticketID)
	uuid := ticketUUID(prefix, ticketID)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		eventType, _ := event["event"].(string)
		ts, _ := event["ts"].(string)
		if eventType == "" || ts == "" {
			continue
		}
		db.InsertBuildEvent(uuid, eventType, ts, line)
	}
}
