#!/usr/bin/env bash
# jab.sh — a tiny LLM workflow in a single file
#
# Usage: ./jab.sh tickets/*
#
# Each file is a "ticket" — a plain-text description of something to build.
# For each ticket, runs four stages:
#   1. Plan     — LLM reads the ticket and produces an implementation plan
#   2. Implement — LLM executes the plan (with tool use)
#   3. Test     — bash command to verify (customize this!)
#   4. Validate — LLM reviews what was built
#
# Requires: claude CLI (https://docs.anthropic.com/en/docs/claude-code)
#
# This is meant to be forked and made yours. Change the prompts,
# swap the test command, add stages, remove stages — go wild.

set -euo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: jab.sh <ticket> [ticket...]"
  exit 1
fi

# --- Customize these prompts ---

plan_prompt() {
  cat <<EOF
Here is a ticket describing work to be done. Produce a short, concrete
implementation plan. List each step as a bullet point.

--- Ticket ---
$(cat "$1")
EOF
}

implement_prompt() {
  cat <<EOF
Here is a ticket and an implementation plan. Do the work described in the plan.

--- Ticket ---
$(cat "$1")

--- Plan ---
$2
EOF
}

validate_prompt() {
  cat <<EOF
A ticket was just implemented. Review the changes and assess whether the
work was completed correctly. Be brief — a few sentences max.

--- Ticket ---
$(cat "$1")

--- Plan ---
$2
EOF
}

# --- Test stage (pure bash — customize this!) ---

run_test() {
  # Replace this with your actual test command, e.g.:
  #   npm run lint && npm test
  #   pytest
  #   go test ./...
  #   cargo test
  echo "No test configured — edit run_test() in this script"
}

# --- LLM backend (uncomment one pair) ---

# Claude Code
# llm()       { echo "$1" | claude -p --output-format text; }
# llm_tools() { echo "$1" | claude -p --output-format text --dangerously-skip-permissions; }

# Codex
# llm()       { echo "$1" | codex --quiet; }
# llm_tools() { echo "$1" | codex --quiet --full-auto; }

# OpenCode
# llm()       { echo "$1" | opencode --pipe; }
# llm_tools() { echo "$1" | opencode --pipe --yes; }

# Cursor (agent mode)
llm()       { echo "$1" | cursor --pipe; }
llm_tools() { echo "$1" | cursor --pipe; }

stage() {
  local name="$1"
  echo ""
  echo "=== $name ==="
}

for ticket in "$@"; do
  echo ""
  echo "━━━ $ticket ━━━"

  stage "Plan"
  plan=$(llm "$(plan_prompt "$ticket")")
  echo "$plan"

  stage "Implement"
  llm_tools "$(implement_prompt "$ticket" "$plan")"

  stage "Test"
  if run_test; then
    echo "PASS"
  else
    echo "FAIL (continuing anyway)"
  fi

  stage "Validate"
  llm "$(validate_prompt "$ticket" "$plan")"
done

echo ""
echo "Done."
