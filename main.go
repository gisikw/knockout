package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		return cmdHelp(nil)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "create":
		return cmdCreate(rest)
	case "show":
		return cmdShow(rest)
	case "status":
		return cmdStatus(rest)
	case "start":
		return cmdStart(rest)
	case "close":
		return cmdClose(rest)
	case "reopen":
		return cmdReopen(rest)
	case "block":
		return cmdBlock(rest)
	case "ls":
		return cmdLs(rest)
	case "ready":
		return cmdReady(rest)
	case "blocked":
		return cmdBlocked(rest)
	case "closed":
		return cmdClosed(rest)
	case "dep":
		return cmdDep(rest)
	case "undep":
		return cmdUndep(rest)
	case "link":
		return cmdLink(rest)
	case "unlink":
		return cmdUnlink(rest)
	case "add-note":
		return cmdAddNote(rest)
	case "query":
		return cmdQuery(rest)
	case "help", "--help", "-h":
		return cmdHelp(rest)
	case "version", "--version", "-v":
		fmt.Println("ko " + version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "ko: unknown command '%s'\n", cmd)
		fmt.Fprintln(os.Stderr, "Run 'ko help' for usage.")
		return 1
	}
}

func cmdHelp(args []string) int {
	fmt.Println(`ko - knockout task tracker

Usage: ko <command> [arguments]

Commands:
  create [title]     Create a new ticket
  show <id>          Show ticket details
  ls                 List open tickets
  ready              Show ready queue (open + deps resolved)
  blocked            Show tickets with unresolved deps
  closed             Show closed tickets

  status <id> <s>    Set ticket status
  start <id>         Set status to in_progress
  close <id>         Set status to closed
  reopen <id>        Set status to open
  block <id>         Set status to blocked

  dep <id> <dep>     Add dependency
  undep <id> <dep>   Remove dependency
  dep tree <id>      Show dependency tree

  link <id1> <id2>   Link two tickets
  unlink <id1> <id2> Unlink two tickets

  add-note <id> <text>  Add a note to a ticket
  query                 Output all tickets as JSONL

  help               Show this help
  version            Show version`)
	return 0
}
