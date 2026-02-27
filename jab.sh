#!/usr/bin/env bash
# jab.sh — a tiny LLM workflow in a single file
#
# Usage: ./jab.sh file1 file2 file3 ...
#
# For each file, runs four stages:
#   1. Plan    — LLM reads the file, proposes changes
#   2. Implement — LLM applies the plan
#   3. Test    — bash command to verify (customize this!)
#   4. Validate — LLM reviews the result
#
# Requires: claude CLI (https://docs.anthropic.com/en/docs/claude-code)
#
# This is meant to be forked and made yours. Change the prompts,
# swap the test command, add stages, remove stages — go wild.

set -euo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: jab.sh <file> [file...]"
  exit 1
fi

# --- Customize these prompts ---

plan_prompt() {
  cat <<EOF
Read this file and propose improvements. Be specific and concise.
List each change as a bullet point.

$(cat "$1")
EOF
}

implement_prompt() {
  cat <<EOF
Here is a file and a plan. Output the complete updated file, nothing else.

--- Plan ---
$2

--- File ---
$(cat "$1")
EOF
}

validate_prompt() {
  cat <<EOF
Review this file. Does it look correct? Note any issues.
Be brief — a few sentences max.

$(cat "$1")
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

# --- Engine (you probably don't need to touch this) ---

llm() {
  echo "$1" | claude -p --output-format text
}

stage() {
  local name="$1"
  echo ""
  echo "=== $name ==="
}

for file in "$@"; do
  echo ""
  echo "━━━ $file ━━━"

  stage "Plan"
  plan=$(llm "$(plan_prompt "$file")")
  echo "$plan"

  stage "Implement"
  result=$(llm "$(implement_prompt "$file" "$plan")")
  echo "$result" > "$file"
  echo "Wrote $file"

  stage "Test"
  if run_test "$file"; then
    echo "PASS"
  else
    echo "FAIL (continuing anyway)"
  fi

  stage "Validate"
  llm "$(validate_prompt "$file")"
done

echo ""
echo "Done."
