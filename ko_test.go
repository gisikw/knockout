package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"ko": func() int { return run(os.Args[1:]) },
	}))
}

// seedTicketsDir imports all .md ticket files from a directory into the DB
// at the given dbPath. No env vars are touched — safe for parallel tests.
func seedTicketsDir(ticketsDir, dbPath string) error {
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
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
	if len(tickets) == 0 {
		return nil
	}

	sort.Slice(tickets, func(i, j int) bool {
		return Depth(tickets[i].ID) < Depth(tickets[j].ID)
	})

	abs, _ := filepath.Abs(ticketsDir)
	db, err := OpenDBAt(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, t := range tickets {
		if err := db.UpsertTicket(t, abs); err != nil {
			return fmt.Errorf("seed %s: %w", t.ID, err)
		}
	}
	return nil
}

// testParams returns testscript.Params with DB isolation configured.
// XDG_STATE_HOME points to a sibling temp dir (not env.WorkDir) so the
// shadow DB never pollutes the project tree — tests that git-init inside
// WorkDir would otherwise see it as an uncommitted change.
// After archive extraction, any .ko/tickets/*.md fixtures are seeded into
// the DB so tests don't depend on filesystem sync.
// A custom "seed" command is available in test scripts to import tickets
// from a directory: `seed <ticketsDir>`.
func testParams(dir string) testscript.Params {
	return testscript.Params{
		Dir: dir,
		Setup: func(env *testscript.Env) error {
			stateDir, err := os.MkdirTemp("", "ko-state-")
			if err != nil {
				return err
			}
			env.Defer(func() { os.RemoveAll(stateDir) })
			env.Setenv("XDG_STATE_HOME", stateDir)
			// Store DB path in env so the seed command can retrieve it per-subtest
			env.Setenv("KO_TEST_DB_PATH", filepath.Join(stateDir, "knockout", "knockout.db"))

			// Seed DB from any ticket fixtures in the archive
			ticketsDir := filepath.Join(env.WorkDir, ".ko", "tickets")
			dbPath := filepath.Join(stateDir, "knockout", "knockout.db")
			if err := seedTicketsDir(ticketsDir, dbPath); err != nil {
				return err
			}
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// seed <ticketsDir> — imports .md ticket files into the DB
			"seed": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) != 1 {
					ts.Fatalf("usage: seed <ticketsDir>")
				}
				dbPath := ts.Getenv("KO_TEST_DB_PATH")
				if dbPath == "" {
					ts.Fatalf("seed: KO_TEST_DB_PATH not set")
				}
				dir := ts.MkAbs(args[0])
				if err := seedTicketsDir(dir, dbPath); err != nil {
					ts.Fatalf("seed: %v", err)
				}
			},
		},
	}
}

func TestTicketCreation(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_creation"))
}

func TestTicketStatus(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_status"))
}

func TestTicketShow(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_show"))
}

func TestTicketListing(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_listing"))
}

func TestTicketDeps(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_deps"))
}

func TestTicketNotes(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_notes"))
}

func TestServe(t *testing.T) {
	testscript.Run(t, testParams("testdata/serve"))
}

func TestIDResolution(t *testing.T) {
	testscript.Run(t, testParams("testdata/id_resolution"))
}

func TestDirectoryResolution(t *testing.T) {
	testscript.Run(t, testParams("testdata/directory_resolution"))
}

func TestPipeline(t *testing.T) {
	testscript.Run(t, testParams("testdata/pipeline"))
}

func TestProjectRegistry(t *testing.T) {
	testscript.Run(t, testParams("testdata/project_registry"))
}

func TestLoop(t *testing.T) {
	testscript.Run(t, testParams("testdata/loop"))
}

func TestAgentHarnesses(t *testing.T) {
	testscript.Run(t, testParams("testdata/agent_harnesses"))
}

func TestTicketSnooze(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_snooze"))
}

func TestTicketTriage(t *testing.T) {
	testscript.Run(t, testParams("testdata/ticket_triage"))
}

func TestAgentTriage(t *testing.T) {
	testscript.Run(t, testParams("testdata/agent_triage"))
}
