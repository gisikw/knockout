package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The shim: a ko-shaped surface that proxies to Questbook's QQL API instead of
// the local store. It is opt-in (KO_QQL=1) and inert by default — nothing about
// normal ko behavior changes until Kevin flips it. Its whole job is to let
// muscle memory and the large body of context-poisoned `ko ...` references keep
// working during and after cutover, while logging every hit so the log going
// quiet is the evidence migration is done. Boring by design; meant to rot away.

// shimEnvVar is the opt-in switch. When set to a non-empty, non-"0"/"false"
// value, ko routes commands through the QQL shim.
const shimEnvVar = "KO_QQL"

// ShimEnabled reports whether the QQL shim is active.
func ShimEnabled() bool {
	v := strings.TrimSpace(os.Getenv(shimEnvVar))
	return v != "" && v != "0" && v != "false"
}

// koToQQLStatus maps ko's status vocabulary onto QQL quest statuses.
var koToQQLStatus = map[string]string{
	"captured":    "open",
	"routed":      "open",
	"open":        "open",
	"in_progress": "in_progress",
	"blocked":     "blocked",
	"closed":      "done",
	"resolved":    "done",
}

// shimCtx carries the QQL client and project mapping through a shim invocation.
type shimCtx struct {
	client *QQLClient
	mapp   *QQLMapping
}

// runShim is the shim entry point, mirroring run()'s dispatch. It logs every
// invocation, then routes to a QQL-backed handler or fails loudly.
func runShim(args []string) int {
	if len(args) == 0 {
		return cmdHelp(nil)
	}
	cmd := args[0]
	rest := args[1:]

	logShimUsage(cmd, rest)

	// Purely local, store-independent commands still work as normal.
	switch cmd {
	case "help", "--help", "-h":
		return cmdHelp(rest)
	case "version", "--version", "-v":
		fmt.Println("ko " + version + " (QQL shim)")
		return 0
	}

	mapp, err := LoadQQLMapping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: qql shim: cannot load mapping: %v\n", err)
		return 1
	}
	s := &shimCtx{client: NewQQLClient(), mapp: mapp}

	switch cmd {
	case "add":
		return s.add(rest)
	case "ls":
		return s.ls(rest)
	case "ready":
		return s.ready(rest)
	case "show":
		return s.show(rest)
	case "dep":
		return s.dep(rest)
	case "undep":
		return s.undep(rest)
	case "status":
		return s.setStatus(rest)
	case "start":
		return s.statusShortcut(rest, "in_progress")
	case "close":
		return s.statusShortcut(rest, "closed")
	case "open":
		return s.statusShortcut(rest, "open")
	case "block":
		return s.statusShortcut(rest, "blocked")
	case "update":
		return s.update(rest)
	case "note":
		return s.note(rest)
	default:
		return shimUnsupported(cmd)
	}
}

// shimUnsupported fails loudly for ko subcommands the shim does not proxy,
// pointing at the QQL equivalent. It NEVER silently no-ops.
func shimUnsupported(cmd string) int {
	pointers := map[string]string{
		"agent":   "pipeline execution is not part of Questbook; run agents against QQL directly (see kobold).",
		"serve":   "the QQL service is `qb serve` in ~/Projects/questbook; ko does not serve QQL.",
		"import":  "ko import populates the legacy store; to load tickets into Questbook use `ko export` + the QQL bulk-import endpoint.",
		"project": "projects map to Questbook realms/campaigns via the mapping file (see QQL_MAPPING.md); manage realms with `qb mutate`.",
		"stats":   "aggregate over QQL with a `qb query` on quests (filter by realm/campaign/status).",
		"search":  "QQL has no full-text search yet; query quests by status/realm/campaign with `qb query`.",
		"history": "quest history lives in Questbook revisions/events; query with `qb query`.",
		"triage":  "triage is not modeled in Questbook; use quest status `blocked` or a question-type quest.",
		"snooze":  "snooze is not modeled in Questbook; park the quest with status or a dependency.",
		"bump":    "priority ordering is the `priority` field; set it via `ko update -p N` (proxied) — there is no mtime bump in QQL.",
		"undep":   "", // handled elsewhere; here for completeness
	}
	msg := pointers[cmd]
	if msg == "" {
		msg = "no QQL equivalent is wired in the shim. Use `qb query`/`qb mutate` directly."
	}
	fmt.Fprintf(os.Stderr, "ko (QQL shim): '%s' is not proxied to Questbook.\n  → %s\n", cmd, msg)
	return 2
}

// --- usage log -------------------------------------------------------------

type shimLogEntry struct {
	Timestamp  string   `json:"ts"`
	Subcommand string   `json:"subcommand"`
	Argv       []string `json:"argv"`
	CWD        string   `json:"cwd,omitempty"`
	Caller     string   `json:"caller,omitempty"`
	QQLURL     string   `json:"qql_url,omitempty"`
}

// shimLogPath returns the usage-log path: KO_SHIM_LOG, else
// $XDG_STATE_HOME/knockout/shim-usage.jsonl (~/.local/state/... by default).
// Deliberately obvious and greppable — the log is the cutover instrument.
func shimLogPath() string {
	if p := os.Getenv("KO_SHIM_LOG"); p != "" {
		return p
	}
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "knockout", "shim-usage.jsonl")
}

// logShimUsage appends one JSONL record per shim invocation. Best-effort: a
// logging failure must never break the command.
func logShimUsage(cmd string, rest []string) {
	path := shimLogPath()
	if path == "" {
		return
	}
	entry := shimLogEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Subcommand: cmd,
		Argv:       sanitizeArgv(rest),
		QQLURL:     NewQQLClient().BaseURL,
	}
	if cwd, err := os.Getwd(); err == nil {
		entry.CWD = cwd
	}
	entry.Caller = callerContext()

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(line, '\n'))
}

// sanitizeArgv caps each argument's length so a giant pasted body doesn't bloat
// the log. Argv is otherwise the user's own ticket text (no secrets), kept for
// forensic value during cutover.
func sanitizeArgv(args []string) []string {
	const cap = 500
	out := make([]string, len(args))
	for i, a := range args {
		if len(a) > cap {
			out[i] = a[:cap] + "…(truncated)"
		} else {
			out[i] = a
		}
	}
	return out
}

// callerContext returns the parent process's command name if cheaply available
// (Linux /proc). Best-effort; empty on failure or other platforms.
func callerContext() string {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return ""
	}
	var ppid string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PPid:") {
			ppid = strings.TrimSpace(strings.TrimPrefix(line, "PPid:"))
			break
		}
	}
	if ppid == "" {
		return ""
	}
	comm, err := os.ReadFile(filepath.Join("/proc", ppid, "comm"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(comm))
}

// --- project / realm resolution -------------------------------------------

// projectTag resolves the ko project tag for this invocation from a --project
// flag, a #tag positional, or (failing those) the cwd via the registry. It
// returns the tag and the remaining args with the project selector removed.
func (s *shimCtx) projectTag(args []string) (string, []string) {
	var tag string
	var remaining []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--project" && i+1 < len(args):
			tag = args[i+1]
			i++
		case strings.HasPrefix(a, "--project="):
			tag = strings.TrimPrefix(a, "--project=")
		case strings.HasPrefix(a, "#"):
			tag = CleanTag(a)
		default:
			remaining = append(remaining, a)
		}
	}
	if tag == "" {
		tag = tagFromCwd()
	}
	return tag, remaining
}

// tagFromCwd reverse-maps the current project root to its registry tag.
func tagFromCwd() string {
	root, err := findProjectRoot()
	if err != nil {
		return ""
	}
	reg, err := LoadRegistry(RegistryPath())
	if err != nil {
		return ""
	}
	for tag, path := range reg.Projects {
		if path == root {
			return tag
		}
	}
	return filepath.Base(root)
}

// resolveRealmID returns the persistent realm id for a slug, creating the realm
// if it does not exist (realms need only slug+name, so this is safe).
func (s *shimCtx) resolveRealmID(slug string) (string, error) {
	resp, err := s.client.Query(map[string]any{
		"entity":  "realm",
		"filters": map[string]any{"slug": slug},
		"return":  []string{"id", "slug"},
	})
	if err != nil {
		return "", err
	}
	if realms := resp.entityList("realms"); len(realms) > 0 {
		if id, _ := realms[0]["id"].(string); id != "" {
			return id, nil
		}
	}
	// Create it.
	// QQL /api/mutate returns id_map natively and rejects unknown top-level
	// keys (only create/update), so we do NOT send a "return" projection.
	resp, err = s.client.Mutate(map[string]any{
		"create": map[string]any{
			"realms": map[string]any{
				"$r": map[string]any{"slug": slug, "name": slug},
			},
		},
	})
	if err != nil {
		return "", err
	}
	if id := resp.IDMap["$r"]; id != "" {
		return id, nil
	}
	return "", fmt.Errorf("could not resolve or create realm %q", slug)
}

// resolveCampaignID returns the campaign id for a slug if it exists. Campaigns
// require a goal to create, so the shim never auto-creates them; an unmapped or
// missing campaign is a warning, not an error (the realm still anchors quests).
func (s *shimCtx) resolveCampaignID(slug string) string {
	if slug == "" {
		return ""
	}
	resp, err := s.client.Query(map[string]any{
		"entity":  "campaign",
		"filters": map[string]any{"slug": slug},
		"return":  []string{"id", "slug"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko (QQL shim): warning: campaign lookup failed for %q: %v\n", slug, err)
		return ""
	}
	if camps := resp.entityList("campaigns"); len(camps) > 0 {
		if id, _ := camps[0]["id"].(string); id != "" {
			return id
		}
	}
	fmt.Fprintf(os.Stderr, "ko (QQL shim): warning: campaign %q not found; quest anchored to realm only\n", slug)
	return ""
}

// --- handlers --------------------------------------------------------------

func (s *shimCtx) add(args []string) int {
	tag, args := s.projectTag(args)
	args = reorderArgs(args, map[string]bool{
		"d": true, "t": true, "p": true, "a": true,
		"parent": true, "external-ref": true, "design": true,
		"acceptance": true, "tags": true, "snooze": true, "triage": true,
	})
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	desc := fs.String("d", "", "description")
	typ := fs.String("t", "", "type")
	priority := fs.Int("p", -1, "priority")
	parent := fs.String("parent", "", "parent quest id")
	design := fs.String("design", "", "design notes")
	acceptance := fs.String("acceptance", "", "acceptance criteria")
	// Accepted-but-ignored ko flags (logged as dropped) so argv shape holds.
	fs.String("a", "", "assignee (not modeled in QQL; ignored)")
	fs.String("external-ref", "", "external ref (ignored)")
	fs.String("tags", "", "tags (not modeled in QQL; ignored)")
	fs.String("snooze", "", "snooze (not modeled in QQL; ignored)")
	fs.String("triage", "", "triage (not modeled in QQL; ignored)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}
	title := "Untitled"
	if fs.NArg() > 0 {
		title = fs.Arg(0)
	}
	body := ""
	if fs.NArg() > 1 {
		body = fs.Arg(1)
	} else if *desc != "" {
		body = *desc
	}
	if *design != "" {
		body += "\n\n## Design\n\n" + *design
	}
	if *acceptance != "" {
		body += "\n\n## Acceptance Criteria\n\n" + *acceptance
	}

	quest := map[string]any{"title": title, "status": "open"}
	if body != "" {
		quest["body"] = body
	}
	if *typ != "" {
		quest["type"] = *typ
	}
	if *priority >= 0 {
		quest["priority"] = *priority
	}

	if *parent != "" {
		quest["parent"] = *parent
	} else {
		realm, campaign := s.mapp.Resolve(tag)
		realmID, err := s.resolveRealmID(realm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
			return 1
		}
		quest["realm"] = realmID
		if campID := s.resolveCampaignID(campaign); campID != "" {
			quest["campaign"] = campID
		}
	}

	// id_map is returned natively; mutate rejects a "return" key (see resolveRealmID).
	resp, err := s.client.Mutate(map[string]any{
		"create": map[string]any{"quests": map[string]any{"$q": quest}},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko add: %v\n", err)
		return 1
	}
	id := resp.IDMap["$q"]
	if id == "" {
		fmt.Fprintln(os.Stderr, "ko add: quest created but no id returned")
		return 1
	}
	fmt.Println(id)
	return 0
}

func (s *shimCtx) ls(args []string) int {
	tag, args := s.projectTag(args)
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	statusFilter := fs.String("status", "", "filter by status")
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Bool("all", false, "include done quests")
	_ = fs.Int("limit", 0, "limit")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"status": true, "limit": true})); err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}

	realm, _ := s.mapp.Resolve(tag)
	realmID, err := s.resolveRealmID(realm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}
	filters := map[string]any{"realm": realmID}
	if *statusFilter != "" {
		filters["status"] = mapKoStatus(*statusFilter)
	}
	resp, err := s.client.Query(map[string]any{
		"entity":  "quest",
		"filters": filters,
		"return":  []string{"id", "title", "status", "priority"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ls: %v\n", err)
		return 1
	}
	quests := resp.entityList("quests")
	sortQuests(quests)
	if *jsonOut {
		emitJSON(quests)
		return 0
	}
	for _, q := range quests {
		fmt.Println(formatQuestLine(q))
	}
	return 0
}

func (s *shimCtx) ready(args []string) int {
	tag, args := s.projectTag(args)
	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Int("limit", 0, "limit")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"limit": true})); err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}
	realm, _ := s.mapp.Resolve(tag)
	realmID, err := s.resolveRealmID(realm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}
	// QQL has no ready-queue view, and its query API does not expose a quest's
	// dependencies (project() drops relation fields), so true dep-gating is not
	// possible here yet. `ready` therefore degrades to "open quests in realm" —
	// a SUPERSET of the real ready queue. Warn on stderr so this is never a
	// silent lie. Once QQL exposes dependency status in queries, gate here.
	fmt.Fprintln(os.Stderr, "ko ready: note — QQL cannot yet report dependency state; showing all open quests (may include dep-blocked ones)")
	resp, err := s.client.Query(map[string]any{
		"entity":  "quest",
		"filters": map[string]any{"realm": realmID, "status": "open"},
		"return":  []string{"id", "title", "status", "priority"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko ready: %v\n", err)
		return 1
	}
	quests := resp.entityList("quests")
	sortQuests(quests)
	if *jsonOut {
		emitJSON(quests)
		return 0
	}
	for _, q := range quests {
		fmt.Println(formatQuestLine(q))
	}
	return 0
}

func (s *shimCtx) show(args []string) int {
	_, args = s.projectTag(args)
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko show: %v\n", err)
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "ko show: quest ID required")
		return 1
	}
	id := fs.Arg(0)
	// NOTE: QQL's query API returns only scalar quest columns — its project()
	// drops every relation projection (dependencies/subquests/parent.slug). So
	// realm/campaign/parent come back as bare IDs and the dependency graph is
	// not readable here at all (it IS stored; `ko dep` writes persist). We show
	// what QQL actually returns and never fabricate a deps line. See WORKLOG.md.
	resp, err := s.client.Query(map[string]any{
		"entity":  "quest",
		"filters": map[string]any{"id": id},
		"return":  []string{"id", "title", "body", "status", "type", "priority", "parent", "realm", "campaign"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko show: %v\n", err)
		return 1
	}
	quests := resp.entityList("quests")
	if len(quests) == 0 {
		fmt.Fprintf(os.Stderr, "ko show: quest '%s' not found\n", id)
		return 1
	}
	q := quests[0]
	if *jsonOut {
		emitJSON(q)
		return 0
	}
	fmt.Printf("id: %s\n", strField(q, "id"))
	fmt.Printf("status: %s\n", strField(q, "status"))
	if t := strField(q, "type"); t != "" {
		fmt.Printf("type: %s\n", t)
	}
	fmt.Printf("priority: %s\n", priorityStr(q))
	if p := strField(q, "parent"); p != "" {
		fmt.Printf("parent: %s\n", p)
	}
	if r := strField(q, "realm"); r != "" {
		fmt.Printf("realm: %s\n", r)
	}
	if c := strField(q, "campaign"); c != "" {
		fmt.Printf("campaign: %s\n", c)
	}
	fmt.Println()
	fmt.Printf("# %s\n", strField(q, "title"))
	if body := strField(q, "body"); body != "" {
		fmt.Println()
		fmt.Print(body)
		if !strings.HasSuffix(body, "\n") {
			fmt.Println()
		}
	}
	return 0
}

func (s *shimCtx) dep(args []string) int {
	if len(args) >= 1 && args[0] == "tree" {
		return s.depTree(args[1:])
	}
	_, args = s.projectTag(args)
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko dep: two quest IDs required")
		fmt.Fprintln(os.Stderr, "usage: ko dep <quest> <dep>  |  ko dep tree <quest>")
		return 1
	}
	quest, dep := args[0], args[1]
	if quest == dep {
		fmt.Fprintln(os.Stderr, "ko dep: quest cannot depend on itself")
		return 1
	}
	_, err := s.client.Mutate(map[string]any{
		"update": map[string]any{
			"quests": map[string]any{
				quest: map[string]any{"dependencies+": []string{dep}},
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko dep: %v\n", err)
		return 1
	}
	fmt.Printf("Added dependency: %s -> %s\n", quest, dep)
	return 0
}

func (s *shimCtx) undep(args []string) int {
	_, args = s.projectTag(args)
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko undep: two quest IDs required")
		return 1
	}
	quest, dep := args[0], args[1]
	_, err := s.client.Mutate(map[string]any{
		"update": map[string]any{
			"quests": map[string]any{
				quest: map[string]any{"dependencies-": []string{dep}},
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko undep: %v\n", err)
		return 1
	}
	fmt.Printf("Removed dependency: %s -> %s\n", quest, dep)
	return 0
}

// depTree fails loudly: the QQL query API does not return a quest's
// dependencies (see show/ready notes), so a dependency tree cannot be walked.
// `ko dep <a> <b>` writes still persist — this is a read-side gap only. Failing
// loud beats printing a misleading single-node "tree" that implies no deps.
func (s *shimCtx) depTree(args []string) int {
	fmt.Fprintln(os.Stderr, "ko (QQL shim): 'dep tree' is not readable via QQL yet.")
	fmt.Fprintln(os.Stderr, "  → The query API returns no quest relations, so the dependency graph can't be walked.")
	fmt.Fprintln(os.Stderr, "  → `ko dep <quest> <dep>` writes ARE stored; inspect the graph with `qb query` once relation projection lands.")
	return 2
}

func (s *shimCtx) setStatus(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko status: quest ID and status required")
		return 1
	}
	return s.applyStatus(args[0], args[1])
}

func (s *shimCtx) statusShortcut(args []string, koStatus string) int {
	_, args = s.projectTag(args)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ko %s: quest ID required\n", koStatus)
		return 1
	}
	return s.applyStatus(args[0], koStatus)
}

func (s *shimCtx) applyStatus(id, koStatus string) int {
	qqlStatus, ok := koToQQLStatus[koStatus]
	if !ok {
		fmt.Fprintf(os.Stderr, "ko: unknown status %q\n", koStatus)
		return 1
	}
	_, err := s.client.Mutate(map[string]any{
		"update": map[string]any{
			"quests": map[string]any{id: map[string]any{"status": qqlStatus}},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko: %v\n", err)
		return 1
	}
	fmt.Printf("%s updated\n", id)
	return 0
}

func (s *shimCtx) update(args []string) int {
	_, args = s.projectTag(args)
	args = reorderArgs(args, map[string]bool{
		"d": true, "t": true, "p": true, "title": true, "status": true,
		"a": true, "parent": true, "tags": true, "external-ref": true,
		"design": true, "acceptance": true, "snooze": true, "triage": true,
	})
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	title := fs.String("title", "", "title")
	desc := fs.String("d", "", "description (appended to body)")
	typ := fs.String("t", "", "type")
	priority := fs.Int("p", -1, "priority")
	status := fs.String("status", "", "status")
	fs.String("a", "", "assignee (ignored)")
	fs.String("tags", "", "tags (ignored)")
	fs.String("parent", "", "parent")
	fs.String("external-ref", "", "external ref (ignored)")
	fs.String("design", "", "design (appended to body)")
	fs.String("acceptance", "", "acceptance (appended to body)")
	fs.String("snooze", "", "snooze (ignored)")
	fs.String("triage", "", "triage (ignored)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "ko update: quest ID required")
		return 1
	}
	id := fs.Arg(0)

	fields := map[string]any{}
	if *title != "" {
		fields["title"] = *title
	}
	if *typ != "" {
		fields["type"] = *typ
	}
	if *priority >= 0 {
		fields["priority"] = *priority
	}
	if *status != "" {
		if !ValidStatus(*status) {
			fmt.Fprintf(os.Stderr, "ko update: invalid status '%s'\nvalid statuses: %s\n", *status, strings.Join(Statuses, " "))
			return 1
		}
		fields["status"] = mapKoStatus(*status)
	}
	if *desc != "" {
		newBody, err := s.appendedBody(id, *desc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
			return 1
		}
		fields["body"] = newBody
	}
	if len(fields) == 0 {
		fmt.Fprintln(os.Stderr, "ko update: no proxyable fields specified to update")
		return 1
	}
	_, err := s.client.Mutate(map[string]any{
		"update": map[string]any{"quests": map[string]any{id: fields}},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko update: %v\n", err)
		return 1
	}
	fmt.Printf("%s updated\n", id)
	return 0
}

func (s *shimCtx) note(args []string) int {
	_, args = s.projectTag(args)
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ko note: quest ID and note text required")
		return 1
	}
	id := args[0]
	note := strings.Join(args[1:], " ")
	ts := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	newBody, err := s.appendedBody(id, fmt.Sprintf("**%s:** %s", ts, note))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}
	_, err = s.client.Mutate(map[string]any{
		"update": map[string]any{"quests": map[string]any{id: map[string]any{"body": newBody}}},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko note: %v\n", err)
		return 1
	}
	fmt.Printf("Note added to %s\n", id)
	return 0
}

// appendedBody fetches a quest's current body and returns it with text appended
// under a Notes section (QQL body updates replace, so we read-modify-write).
func (s *shimCtx) appendedBody(id, text string) (string, error) {
	resp, err := s.client.Query(map[string]any{
		"entity":  "quest",
		"filters": map[string]any{"id": id},
		"return":  []string{"id", "body"},
	})
	if err != nil {
		return "", err
	}
	quests := resp.entityList("quests")
	if len(quests) == 0 {
		return "", fmt.Errorf("quest '%s' not found", id)
	}
	body := strField(quests[0], "body")
	if strings.Contains(body, "## Notes") {
		return body + "\n" + text + "\n", nil
	}
	if body != "" {
		return body + "\n\n## Notes\n\n" + text + "\n", nil
	}
	return "## Notes\n\n" + text + "\n", nil
}

// --- small helpers ---------------------------------------------------------

// mapKoStatus maps a ko status to its QQL equivalent, passing through anything
// already in QQL vocabulary.
func mapKoStatus(status string) string {
	if q, ok := koToQQLStatus[status]; ok {
		return q
	}
	return status
}

func strField(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func priorityStr(q map[string]any) string {
	if v, ok := q["priority"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%d", int(f))
		}
	}
	return "2"
}

func formatQuestLine(q map[string]any) string {
	return fmt.Sprintf("%s [%s] (p%s) %s",
		strField(q, "id"), strField(q, "status"), priorityStr(q), strField(q, "title"))
}

// sortQuests orders quests by priority ascending then id, mirroring ko.
func sortQuests(quests []map[string]any) {
	sort.SliceStable(quests, func(i, j int) bool {
		pi, pj := questPriority(quests[i]), questPriority(quests[j])
		if pi != pj {
			return pi < pj
		}
		return strField(quests[i], "id") < strField(quests[j], "id")
	})
}

func questPriority(q map[string]any) int {
	if f, ok := q["priority"].(float64); ok {
		return int(f)
	}
	return 2
}

func emitJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
