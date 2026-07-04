package main

import (
	"fmt"
	"os"
	"strings"
)

// Read-only legacy mode. Post-cutover, the legacy local-store paths should stop
// accepting writes while still serving reads (for history spelunking). This is
// opt-in via KO_READONLY=1 and independent of the QQL shim: it guards the
// *legacy* store, pointing writers at the shim / QQL instead of silently
// dropping their mutation.

const readonlyEnvVar = "KO_READONLY"

// ReadonlyEnabled reports whether legacy read-only mode is active.
func ReadonlyEnabled() bool {
	v := strings.TrimSpace(os.Getenv(readonlyEnvVar))
	return v != "" && v != "0" && v != "false"
}

// legacyWriteCommands are the ko subcommands that mutate the local ticket store.
var legacyWriteCommands = map[string]bool{
	"add":    true,
	"update": true,
	"status": true,
	"start":  true,
	"close":  true,
	"open":   true,
	"block":  true,
	"snooze": true,
	"dep":    true,
	"undep":  true,
	"note":   true,
	"bump":   true,
}

// isLegacyWrite reports whether (cmd, rest) would write to the legacy store.
// It accounts for read-only sub-modes of otherwise-mutating commands:
//   - `dep tree ...` is a read (dependency inspection)
//   - `triage` with fewer than two args lists triaged tickets (a read); with
//     two or more it sets triage (a write)
func isLegacyWrite(cmd string, rest []string) bool {
	if cmd == "dep" && len(rest) > 0 && rest[0] == "tree" {
		return false
	}
	if cmd == "triage" {
		return len(rest) >= 2
	}
	return legacyWriteCommands[cmd]
}

// rejectLegacyWrite prints a loud, actionable rejection for a blocked write.
func rejectLegacyWrite(cmd string) int {
	fmt.Fprintf(os.Stderr, "ko: legacy store is read-only (%s set).\n", readonlyEnvVar)
	fmt.Fprintf(os.Stderr, "  '%s' would write to the local ko store, which has been frozen for the Questbook cutover.\n", cmd)
	fmt.Fprintln(os.Stderr, "  → Write via the QQL shim instead: set KO_QQL=1 (proxies to Questbook), or use `qb mutate`.")
	fmt.Fprintln(os.Stderr, "  Reads (show, ls, ready, search, stats, history, dep tree) still work.")
	return 3
}
