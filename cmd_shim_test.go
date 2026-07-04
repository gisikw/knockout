package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShimEnabled(t *testing.T) {
	orig := os.Getenv(shimEnvVar)
	defer os.Setenv(shimEnvVar, orig)
	cases := map[string]bool{"": false, "0": false, "false": false, "1": true, "yes": true}
	for val, want := range cases {
		os.Setenv(shimEnvVar, val)
		if got := ShimEnabled(); got != want {
			t.Errorf("ShimEnabled(%q) = %v, want %v", val, got, want)
		}
	}
}

func TestMapKoStatus(t *testing.T) {
	cases := map[string]string{
		"open":        "open",
		"captured":    "open",
		"routed":      "open",
		"in_progress": "in_progress",
		"blocked":     "blocked",
		"closed":      "done",
		"resolved":    "done",
		"done":        "done", // already-QQL vocabulary passes through
	}
	for ko, want := range cases {
		if got := mapKoStatus(ko); got != want {
			t.Errorf("mapKoStatus(%q) = %q, want %q", ko, got, want)
		}
	}
}

func TestSanitizeArgv(t *testing.T) {
	long := strings.Repeat("x", 600)
	out := sanitizeArgv([]string{"short", long})
	if out[0] != "short" {
		t.Errorf("short arg changed: %q", out[0])
	}
	if !strings.HasSuffix(out[1], "…(truncated)") || len(out[1]) > 600 {
		t.Errorf("long arg not truncated: len=%d", len(out[1]))
	}
}

func TestFormatQuestLineAndPriority(t *testing.T) {
	q := map[string]any{"id": "q-1", "status": "open", "priority": float64(1), "title": "hello"}
	if got := formatQuestLine(q); got != "q-1 [open] (p1) hello" {
		t.Errorf("formatQuestLine = %q", got)
	}
	// Missing priority defaults to 2 (ko default).
	q2 := map[string]any{"id": "q-2", "status": "open", "title": "no prio"}
	if got := priorityStr(q2); got != "2" {
		t.Errorf("priorityStr default = %q, want 2", got)
	}
}

func TestSortQuests(t *testing.T) {
	quests := []map[string]any{
		{"id": "q-c", "priority": float64(2)},
		{"id": "q-a", "priority": float64(1)},
		{"id": "q-b", "priority": float64(1)},
	}
	sortQuests(quests)
	got := []string{strField(quests[0], "id"), strField(quests[1], "id"), strField(quests[2], "id")}
	want := []string{"q-a", "q-b", "q-c"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortQuests order = %v, want %v", got, want)
		}
	}
}

func TestShimUnsupportedFailsLoud(t *testing.T) {
	// Every unsupported command must return a nonzero code and print a pointer.
	for _, cmd := range []string{"agent", "serve", "import", "stats", "search", "history", "snooze", "bump"} {
		code := shimUnsupported(cmd)
		if code == 0 {
			t.Errorf("shimUnsupported(%q) returned 0, want nonzero", cmd)
		}
	}
}

// fakeQuests is the fixed quest set the fake QQL server filters over. Each has a
// ko-namespaced external_ref so external_ref resolution (A1) is exercised, and a
// mix of statuses so the default non-terminal ls filter (A2) is exercised.
var fakeQuests = []map[string]any{
	{"id": "q-1", "external_ref": "ko:tmt-1", "title": "first", "status": "open", "priority": float64(1)},
	{"id": "q-2", "external_ref": "ko:tmt-2", "title": "second", "status": "open", "priority": float64(2)},
	{"id": "q-done", "external_ref": "ko:old-9", "title": "finished", "status": "done", "priority": float64(3)},
}

// fakeQQL is a minimal stand-in for the Questbook QQL API for round-trip tests.
// Its /api/query respects the id, external_ref, and status quest filters so the
// shim's resolution and filtering logic is genuinely tested (realm is treated as
// "all quests are in the realm").
func fakeQQL(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		entity, _ := req["entity"].(string)
		filters, _ := req["filters"].(map[string]any)
		switch entity {
		case "realm":
			writeJSON(w, QQLResponse{Entities: map[string]any{
				"realms": []any{map[string]any{"id": "r-fake", "slug": "testrealm"}},
			}})
		case "quest":
			var matched []any
			for _, q := range fakeQuests {
				if fakeQuestMatches(q, filters) {
					matched = append(matched, q)
				}
			}
			writeJSON(w, QQLResponse{Entities: map[string]any{"quests": matched}})
		default:
			writeJSON(w, QQLResponse{})
		}
	})
	mux.HandleFunc("/api/mutate", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		if _, ok := req["create"]; ok {
			writeJSON(w, QQLResponse{IDMap: map[string]string{"$q": "q-created"}})
			return
		}
		writeJSON(w, QQLResponse{Entities: map[string]any{"ok": true}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// fakeQuestMatches applies the id/external_ref/status exact-match filters the
// shim uses; realm (and any unknown filter) is treated as "matches all".
func fakeQuestMatches(q map[string]any, filters map[string]any) bool {
	for k, v := range filters {
		switch k {
		case "id", "external_ref", "status", "title":
			if q[k] != v {
				return false
			}
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, resp QQLResponse) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// shimTestEnv points the shim at a fake server, an empty (but present) mapping,
// and a temp log. The mapping file must exist: an explicitly-set-but-missing
// KO_QQL_MAPPING is now a loud error (A3), so tests write a real empty file.
func shimTestEnv(t *testing.T, url string) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "usage.jsonl")
	mappingPath := filepath.Join(t.TempDir(), "mapping.yaml")
	if err := os.WriteFile(mappingPath, []byte("default_realm: testrealm\n"), 0644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}
	t.Setenv("KO_QQL", "1")
	t.Setenv("KO_QQL_URL", url)
	t.Setenv("KO_QQL_MAPPING", mappingPath)
	t.Setenv("KO_SHIM_LOG", logPath)
	return logPath
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestShimAddRoundTrip(t *testing.T) {
	srv := fakeQQL(t)
	logPath := shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"add", "--project=testrealm", "A new quest"}); code != 0 {
			t.Errorf("runShim add returned %d", code)
		}
	})
	if strings.TrimSpace(out) != "q-created" {
		t.Errorf("add printed %q, want q-created", strings.TrimSpace(out))
	}

	// Usage must be logged as JSONL.
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage log not written: %v", err)
	}
	var entry shimLogEntry
	if err := json.Unmarshal([]byte(strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)[0]), &entry); err != nil {
		t.Fatalf("usage log not valid JSONL: %v", err)
	}
	if entry.Subcommand != "add" {
		t.Errorf("logged subcommand = %q, want add", entry.Subcommand)
	}
}

func TestShimLsRoundTrip(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"ls", "--project=testrealm"}); code != 0 {
			t.Errorf("runShim ls returned %d", code)
		}
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("ls printed %d lines, want 2: %q", len(lines), out)
	}
	// Sorted by priority: q-1 (p1) before q-2 (p2).
	if !strings.HasPrefix(lines[0], "q-1 [open] (p1)") {
		t.Errorf("first line = %q, want q-1 p1 first", lines[0])
	}
}

func TestShimStatusRoundTrip(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"close", "q-1"}); code != 0 {
			t.Errorf("runShim close returned %d", code)
		}
	})
	if strings.TrimSpace(out) != "q-1 updated" {
		t.Errorf("close printed %q, want 'q-1 updated'", strings.TrimSpace(out))
	}
}

func TestShimUnsupportedViaRunShim(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)
	if code := runShim([]string{"stats"}); code == 0 {
		t.Error("runShim stats returned 0, want nonzero (unsupported)")
	}
}

// A1: a legacy ko-style ref resolves via external_ref to the qb id.
func TestShimResolvesKoRef(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"show", "tmt-1"}); code != 0 {
			t.Errorf("runShim show tmt-1 returned %d", code)
		}
	})
	if !strings.Contains(out, "id: q-1") {
		t.Errorf("show tmt-1 did not resolve to q-1:\n%s", out)
	}
}

// A1: a native qb id still resolves directly (no regression).
func TestShimResolvesQBID(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"show", "q-2"}); code != 0 {
			t.Errorf("runShim show q-2 returned %d", code)
		}
	})
	if !strings.Contains(out, "id: q-2") {
		t.Errorf("show q-2 output missing id: q-2:\n%s", out)
	}
}

// A1: an unknown ref fails loudly, naming both lookups it tried.
func TestShimUnknownRefFailsLoud(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	stderr := captureStderr(t, func() {
		if code := runShim([]string{"show", "nope-0000"}); code == 0 {
			t.Error("runShim show nope-0000 returned 0, want nonzero")
		}
	})
	if !strings.Contains(stderr, "external_ref ko:nope-0000") {
		t.Errorf("not-found error did not name both lookups:\n%s", stderr)
	}
}

// A1: status shortcuts resolve ko refs too, and echo the resolved qb id.
func TestShimStatusResolvesKoRef(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	out := captureStdout(t, func() {
		if code := runShim([]string{"close", "tmt-2"}); code != 0 {
			t.Errorf("runShim close tmt-2 returned %d", code)
		}
	})
	if strings.TrimSpace(out) != "q-2 updated" {
		t.Errorf("close tmt-2 printed %q, want 'q-2 updated'", strings.TrimSpace(out))
	}
}

// A2: `ko ls` hides done quests by default; `--all` includes them.
func TestShimLsHidesDoneByDefault(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)

	def := captureStdout(t, func() {
		if code := runShim([]string{"ls"}); code != 0 {
			t.Errorf("runShim ls returned %d", code)
		}
	})
	if strings.Contains(def, "q-done") {
		t.Errorf("ls (default) leaked a done quest:\n%s", def)
	}

	all := captureStdout(t, func() {
		if code := runShim([]string{"ls", "--all"}); code != 0 {
			t.Errorf("runShim ls --all returned %d", code)
		}
	})
	if !strings.Contains(all, "q-done") {
		t.Errorf("ls --all omitted the done quest:\n%s", all)
	}
}

// A3: an explicitly-set-but-missing mapping file is a loud error, not a silent
// empty tracker.
func TestShimMissingMappingErrorsLoud(t *testing.T) {
	srv := fakeQQL(t)
	shimTestEnv(t, srv.URL)
	// Override the mapping to a path that does not exist.
	t.Setenv("KO_QQL_MAPPING", filepath.Join(t.TempDir(), "does-not-exist.yaml"))

	if code := runShim([]string{"ls"}); code == 0 {
		t.Error("runShim ls with missing explicit mapping returned 0, want nonzero")
	}
}

// captureStderr mirrors captureStdout for error-path assertions.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = orig }()
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}
