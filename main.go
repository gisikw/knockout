package main

import (
	"fmt"
	"os"
	"strings"
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
	case "build":
		return cmdBuild(rest)
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

// reorderArgs moves flags before positional arguments so that Go's flag
// package can parse them regardless of where the caller placed them.
// valueFlags is the set of flag names (without leading dash) that consume
// the next argument as a value (e.g. "p", "d", "parent").
func reorderArgs(args []string, valueFlags map[string]bool) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			// Check if this flag consumes the next arg as a value.
			// Handles both "-p 0" and "--parent ko-a001" forms.
			// Does NOT consume next arg if the flag contains "=" (e.g. "--status=open").
			if !strings.Contains(a, "=") {
				name := strings.TrimLeft(a, "-")
				if valueFlags[name] && i+1 < len(args) {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			positional = append(positional, a)
		}
	}
	return append(flags, positional...)
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

  build <id>         Run build pipeline against ticket

  help               Show this help
  version            Show version`)
	return 0
}
