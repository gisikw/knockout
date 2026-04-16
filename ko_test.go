package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"ko": func() int { return run(os.Args[1:]) },
	}))
}

// testParams returns testscript.Params with DB isolation configured.
// Each test gets its own SQLite database via XDG_STATE_HOME=$WORK.
func testParams(dir string) testscript.Params {
	return testscript.Params{
		Dir: dir,
		Setup: func(env *testscript.Env) error {
			// Point DB to test's work directory for isolation
			env.Setenv("XDG_STATE_HOME", env.WorkDir)
			return nil
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
