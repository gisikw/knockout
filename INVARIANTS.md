# Invariants

Explicit architectural and taste decisions for the knockout codebase.
These are contracts, not suggestions. Violating an invariant is a bug.
If an invariant is wrong, update the invariant first — then change the code.

These invariants reflect engineering judgment about how this codebase should
work. They will sometimes be wrong — that's the point of making them explicit.
Challenge them by updating them, not by silently ignoring them.

Existing code that violates these invariants is out of compliance and should
be ticketed for remediation. No grandfathering.

## Specifications and Tests

- **Every behavior has a spec.** Behavioral specs live in `specs/*.feature`
  (gherkin syntax). These are the source of truth for what the system promises
  to do. They are documentation artifacts, not executable test suites (though
  they may become executable).
- **Every spec has a test.** Go tests using `testscript` in `testdata/*.txtar`
  files are the verification layer. A spec without a corresponding test is an
  unverified claim. A test without a corresponding spec is a test that can be
  silently removed — there's no way to know if it's validating the right thing.
- **Specs and tests are independent artifacts.** A discrepancy between a spec
  and its corresponding test (missing test, skipped test, test that doesn't
  actually validate the spec's claim) is always a defect — the question is
  which one is wrong.
- **Spec before code.** Every new behavior gets a spec before or alongside the
  implementation. Not after. The spec is how we know what we're building.
  Writing the test first is fine; writing the code first and speccing it later
  is how intent gets lost.

## Build

- **Version is injected via ldflags from `git rev-parse --short HEAD`.** No
  hardcoded version strings.
- **Zero external runtime dependencies.** `ko` is a single static binary. It
  must not shell out to other tools (no `git`, `jq`, `awk`, `sed` at runtime).
  The bash implementation relied on coreutils; the Go implementation does not.
  Build-time dependencies (Go toolchain) are fine. Test-time dependencies
  (testscript) are fine. Runtime: just the binary.

## Data Model

- **Tickets are markdown files with YAML frontmatter.** The file is the source
  of truth. There is no database, no index, no derived state that can drift.
  If the file says it, that's what it is.
- **Ticket IDs encode hierarchy.** Root tickets: `<prefix>-<hash>`. Children:
  `<parent-id>.<hash>`. Depth is visible in the ID (count the dots). This is
  a structural invariant, not cosmetic — decomposition depth limits depend on it.
- **The project registry is a file, not a convention.** Cross-project routing
  and dep resolution require knowing where projects live. This mapping is
  explicit and lives in a config file, not inferred from directory structure.
- **Statuses are a closed set.** `captured | routed | open | in_progress |
  closed | blocked`. No other statuses. `ready` is a computed query (open +
  deps resolved), not a status.
- **`ready` never returns tickets in `captured`, `routed`, or `blocked`
  status.** Only `open` and `in_progress` with all deps resolved.

## Pipeline (build subsystem)

- **Workflows, not stages.** The pipeline config declares named `workflows`,
  each containing typed nodes. Every ticket enters the `main` workflow.
  Reaching the end of a workflow = succeed (close the ticket).
- **Two node types: decision and action.** Decision nodes return structured
  JSON dispositions (extracted from the last fenced code block). Action nodes
  just do work — their output is not parsed. This separation eliminates the
  fragile first-line parsing of v1.
- **Dispositions are a closed set.** `continue | fail | blocked | decompose |
  route`. Decision nodes must return one of these. Unknown dispositions are
  errors.
- **Route targets must be declared.** A decision node's `routes` field lists
  the workflows it can jump to. Attempting to route to an undeclared target
  is a build failure, not a redirect. This prevents prompt injection from
  hijacking the workflow graph.
- **Node visits are bounded.** Each node has a `max_visits` (default 1) that
  caps how many times it can be entered during a single build. Exceeding the
  limit is a build failure (ticket blocked). This bounds cycles — routing to
  self is allowed, but only within the visit limit.
- **Every build outcome removes the ticket from `ready`.** SUCCEED closes.
  FAIL marks blocked. BLOCKED wires a dep. DECOMPOSE creates children and
  blocks parent on them. No outcome leaves a ticket sitting on the ready
  queue unchanged. This is the loop-safety invariant.
- **Decomposition is depth-bounded.** The maximum decomposition depth is
  enforced by counting dots in the ticket ID. At the limit, the outcome is
  FAIL (needs human), not DECOMPOSE. This prevents runaway work generation.
- **At most one external ask per build.** A build run can create unbounded
  local subtasks (within depth limits) but at most one cross-project ask.
  It blocks on that ask. This bounds blast radius.
- **Workspace persists across the build.** Each build creates a workspace at
  `.ko/tickets/<id>.artifacts/workspace/`. Stage outputs are tee'd to
  `<workflow>.<node>.md`. Exposed as `$KO_TICKET_WORKSPACE` (the workspace
  path) and `$KO_ARTIFACT_DIR` (the parent artifact directory). The artifact
  directory persists across builds and is cleaned on ticket close. This
  replaces single-stage-back output threading.
- **Invalid disposition JSON is retry-eligible.** If a decision node produces
  output without a valid fenced JSON block, or the JSON doesn't parse, the
  node is retried (up to `max_retries`). Valid dispositions (even `fail`) are
  never retried.
- **`on_close` runs after the ticket is closed.** If an `on_close` command
  kills the process (e.g. service restart), the ticket is already closed.
  This prevents deploy loops.
- **Loop mode prevents ticket creation.** When `ko loop` is running, it sets
  `KO_NO_CREATE=1`. Both `ko create` and `ko add` refuse to execute when
  this variable is set. This is a hard gate — spawned agents cannot create
  tickets during a loop, preventing runaway scope expansion.

## Code Organization

- **Decision logic is pure.** Functions that make decisions (is this ticket
  ready? should we decompose? what outcome does this signal?) take data in
  and return decisions out. No file I/O, no exec calls.
- **I/O is plumbing, not logic.** Functions that read tickets, write files,
  or shell out to build stages are thin orchestrators: gather data, call pure
  decision functions, act on results.
- **New logic goes into testable functions first.** If the first thing you
  write is file I/O, you're doing it backwards. Write the decision function,
  write the test, then wire it into the CLI.
- **No multi-purpose functions.** A function that decides *and* acts is doing
  two things. Separate the decision from the effect.

## File Size

- **500 lines max per file.** This is an ergonomic constraint, not an aesthetic
  one. Every edit requires a preceding read. At 2500 lines, a single task burns
  5+ partial reads just to orient — that's context window spent on navigation
  instead of reasoning. 500 lines fits in one read and leaves room to think.
- **Split along behavioral seams, not alphabetically.** A file should be one
  coherent unit: ticket CRUD, dep resolution, pipeline execution, registry
  lookups. Not "functions A-M" and "functions N-Z".
- **Tests mirror source files.** `ticket.go` → `ticket_test.go`. Shared test
  infrastructure (mocks, factories, helpers) lives in `testutil_test.go`.
- **`main.go` is just `main()`.** Flag parsing, subcommand dispatch. No
  business logic, no method definitions.
- **No `util.go`.** If a function is useful, it belongs with the domain
  that uses it. A `util.go` is a smell.
- **Existing files over 500 lines are out of compliance.** Ticket the split,
  don't let new work make them bigger.

## Error Handling

- **CLI errors go to stderr with a non-zero exit code.** Structured enough
  for scripts to parse, human-readable enough for interactive use.
- **Fail fast on bad input.** Invalid ticket IDs, missing files, malformed
  YAML — these are immediate errors, not things to recover from. The user
  re-runs with correct input.
- **Pipeline failures are outcomes, not crashes.** A stage failing is a
  FAIL outcome, not a panic. The build runner always completes and reports.
  The only crashes are genuine bugs (nil pointers, impossible states).

## Naming

- **Specs are named for the behavioral domain, not the implementation.**
  `ticket_creation.feature`, not `createTicket_test.feature`. The spec
  describes what the system does, not which function does it.
- **One spec file per behavioral domain.** Don't split a domain across files.
  Don't combine unrelated domains.
- **Timestamps in filenames use `2006-01-02_15-04-05` format.** Sorts
  lexicographically, filesystem-safe. No colons, no spaces.

## Policy

- **Decisions that shape code are explicit, not implicit.** If an agent would
  need to read three files and infer a pattern, that pattern should be
  documented here instead. Code is evidence of decisions; this file is where
  the decisions themselves live.
- **No implicit patterns.** "Look at how the other files do it" is not a
  policy. If a convention matters, it's written here. If it's not written
  here, it's not a convention — it's a coincidence.
