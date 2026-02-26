# Research Findings: Should we add behave to flake.nix and integrate specs into test suite?

## Executive Summary

**Recommendation: No, do not add behave or Python dependencies.**

The `.feature` files in `specs/` are already fully covered by Go's testscript-based txtar tests in `testdata/`. Adding behave would introduce:
- Language duplication (Python + Go)
- Additional dependencies without benefit
- Maintenance overhead for parallel test suites

The project already has a complete, working test suite. The feature files serve better as **documentation** than executable tests.

---

## Current State

### Test Infrastructure

The project uses **Go's testscript framework** (github.com/rogpeppe/go-internal/testscript v1.14.1) for all functional testing:

- **98 txtar test files** across multiple domains
- **Test organization:** `testdata/<domain>/*.txtar` files
- **Test runner:** `ko_test.go` with 13 test functions
- **Test execution:** Standard `go test ./...` (no extra setup needed)

Example test structure (testdata/ticket_creation/create_basic.txtar):
```txtar
# Create a basic ticket with title
exec ko add 'My first ticket'
stdout '^ko-[0-9a-f]{4}$'
```

### Feature Files Status

The project has **16 .feature files** (273 scenarios, ~2,091 lines) covering:
- ticket_creation.feature (28 scenarios)
- pipeline.feature (54 scenarios)
- http_server.feature (22 scenarios)
- ticket_listing.feature (24 scenarios)
- ticket_status.feature (31 scenarios)
- And 11 more feature files

**Critical finding:** There are **no Python step definitions** (no `specs/steps/` directory). The feature files are Gherkin syntax only.

### Test Coverage Overlap Analysis

I compared `specs/ticket_creation.feature` with `testdata/ticket_creation/*.txtar`:

**Feature file scenarios:**
- "Create a basic ticket with title"
- "Create a ticket with default title"
- "Create a ticket with description using -d flag"
- "Create with type/priority/assignee"
- etc.

**Corresponding txtar tests:**
- `create_basic.txtar` - covers basic creation + default title
- `create_options.txtar` - covers -d, -t, -p, -a flags
- `create_defaults.txtar` - covers default values
- `create_hierarchy.txtar` - covers parent relationships
- `create_prefix_*.txtar` - covers prefix derivation rules

**Coverage assessment:** The txtar tests comprehensively implement what the feature files describe, often with MORE detail (e.g., testing exact JSON output, regex patterns, edge cases).

---

## Technical Context

### Python/behave Not Present

Current environment:
- `python3`: Not found in PATH
- `behave`: Not installed
- No `flake.nix` in this project
- This is a Go monorepo (go.mod, all tests in Go)

### Related Tickets

**ko-3639** (closed, 2026-02-24): "Make specs runnable with behave"

The ticket notes:
> "Introduce behave (Python BDD framework) as a test dependency and implement step definitions so the specs actually execute."
>
> "Alternatively, evaluate whether the txtar-based tests already cover the same ground and the .feature files are redundant. If so, pick one approach and commit to it."

**Status:** Closed with SUCCEED disposition - but no implementation was added. The ticket was investigating the question, not executing it.

### Why Feature Files Exist

Based on git history and file timestamps, the feature files appear to be:
1. **Design artifacts** - written during feature planning
2. **Documentation** - human-readable behavior specs
3. **Legacy** - possibly from an earlier design phase before testscript was adopted

They serve as **living documentation** of intended behavior, which has value independent of being executable.

---

## Analysis: Why NOT to Add behave

### 1. No Gap in Test Coverage

The txtar tests already provide:
- ✅ 98 test files covering all major features
- ✅ Integration testing with real `ko` binary
- ✅ File system state verification
- ✅ Command line flag combinations
- ✅ Error condition handling
- ✅ Pipeline execution scenarios

Adding behave would duplicate existing coverage, not extend it.

### 2. Language & Toolchain Fragmentation

Current: **Single language (Go)**
- Build: `go build`
- Test: `go test ./...`
- CI: Simple, no multi-language coordination

With behave: **Two languages (Go + Python)**
- Need Python runtime
- Need pip/poetry/virtualenv setup
- Need `behave` + dependencies
- Need to run two test suites: `go test` AND `behave`
- Need to maintain step definitions in Python that call Go binary

### 3. flake.nix Question

The ticket asks about adding dependencies to flake.nix, but:
- **This project has no flake.nix** (it's a Go project)
- Other projects in the homelab have flakes (`/home/dev/Projects/fort-nix/flake.nix`)
- Adding a flake.nix just for behave would be creating infrastructure solely to support redundant tests

If this were intended to be Nix-packaged, you would create a flake.nix that:
```nix
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }: {
    packages.x86_64-linux.default =
      let pkgs = nixpkgs.legacyPackages.x86_64-linux;
      in pkgs.buildGoModule { ... };

    devShells.x86_64-linux.default = pkgs.mkShell {
      buildInputs = [
        pkgs.go
        pkgs.just
        # If behave were added:
        (pkgs.python3.withPackages (ps: [ ps.behave ]))
      ];
    };
  };
}
```

But this is only justified if there's a clear benefit, which there isn't.

### 4. Maintenance Burden

With behave:
- Two test suites to keep in sync
- Two languages of test code to maintain
- Step definitions that parse English sentences (fragile)
- More CI complexity and runtime

Example step definition overhead:
```python
# specs/steps/ticket_steps.py
@when('I run "ko add \'{title}\'"')
def step_run_ko_add(context, title):
    context.result = subprocess.run(['ko', 'add', title], ...)

@then('the output should match a ticket ID pattern')
def step_check_id_pattern(context):
    assert re.match(r'^ko-[0-9a-f]{4}$', context.result.stdout)
```

Compare to txtar (self-contained, declarative):
```txtar
exec ko add 'My first ticket'
stdout '^ko-[0-9a-f]{4}$'
```

The txtar approach is simpler and more maintainable.

### 5. testscript Advantages

The testscript framework provides:
- **Hermetic environments** - each test gets clean state
- **Inline test data** - txtar format embeds file trees
- **Shell-like DSL** - familiar to CLI tool developers
- **Fast execution** - in-process, no subprocess overhead
- **Single binary** - no external dependencies

These advantages would be lost with behave's subprocess-based approach.

---

## Alternative Approach: Use Feature Files as Docs

Rather than executing the feature files, leverage them for their documentation value:

1. **Keep feature files as specification documents**
   - Human-readable behavior descriptions
   - Product requirements in Gherkin format
   - Reference during code review

2. **Generate coverage reports**
   - Script to map feature scenarios → txtar tests
   - Identify any gaps (though analysis suggests full coverage)

3. **Consider auto-generating test stubs**
   - Parse feature files → generate txtar template skeletons
   - Developer fills in the txtar test details
   - Keeps specs and tests in sync at creation time

4. **Link in test comments**
   - Add comments to txtar tests referencing feature file scenarios
   - Example: `# Implements: ticket_creation.feature:9 "Create a basic ticket with title"`

---

## Recommended Actions

### Immediate (No Code Changes)

1. **Document the decision** - Update README.md to clarify:
   - `specs/*.feature` are **design documentation**, not executable tests
   - Actual tests are in `testdata/` using Go testscript
   - To run tests: `go test ./...` or `just test`

2. **Close or update ko-3639** - The ticket asked to evaluate, and evaluation is complete

3. **Archive or preserve feature files** - They have value as:
   - Historical record of design decisions
   - Human-readable requirements
   - Onboarding documentation for new contributors

### Future (If Gaps Found)

If specific scenarios in `.feature` files are NOT covered by txtar tests:
- Add those scenarios as NEW txtar tests (not via behave)
- Keep the language/toolchain uniform

### Do NOT Do

- ❌ Add Python/behave dependencies
- ❌ Create flake.nix solely for behave
- ❌ Write Python step definitions
- ❌ Maintain parallel test suites
- ❌ Introduce multi-language testing complexity

---

## Conclusion

The feature files and txtar tests tell a story of **evolutionary design**:
1. Feature files were written first (planning phase)
2. Comprehensive txtar tests were implemented
3. Feature files remained as documentation

This is a **good outcome**. The project has:
- ✅ High-quality, maintainable Go tests
- ✅ Readable specification documents
- ✅ Simple toolchain (Go only)
- ✅ Fast test execution

Adding behave would move backwards from this state without adding value.

**Final recommendation: Do NOT add behave. Keep feature files as documentation, maintain txtar tests as the single source of test truth.**
