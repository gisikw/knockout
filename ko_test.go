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

func TestTicketCreation(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_creation"})
}

func TestTicketStatus(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_status"})
}

func TestTicketShow(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_show"})
}

func TestTicketListing(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_listing"})
}

func TestTicketDeps(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_deps"})
}

func TestTicketNotes(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_notes"})
}

func TestTicketQuery(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/ticket_query"})
}

func TestIDResolution(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/id_resolution"})
}

func TestDirectoryResolution(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/directory_resolution"})
}

func TestPipeline(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/pipeline"})
}

func TestProjectRegistry(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/project_registry"})
}

func TestLoop(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "testdata/loop"})
}
