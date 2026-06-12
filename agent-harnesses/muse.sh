#!/bin/sh
# Built-in harness for Muse agent CLI
# Receives parameters via KO_-namespaced environment variables

set -e

# Build command arguments
args=""

# Read-only mode maps to --read-only (no write/edit/bash tools)
# Allow-all has no direct equivalent — muse has no permission prompt system
# KO_ALLOW_ALL is effectively a no-op since muse doesn't gate tools behind prompts

if [ -n "$KO_MODEL" ]; then
  args="$args --model $KO_MODEL"
fi

if [ -n "$KO_SYSTEM_PROMPT" ]; then
  # Muse takes --system-prompt-file (path), but KO_SYSTEM_PROMPT is the text.
  # Write to a temp file and clean up on exit.
  _ko_sp=$(mktemp)
  trap 'rm -f "$_ko_sp"' EXIT
  printf '%s' "$KO_SYSTEM_PROMPT" > "$_ko_sp"
  args="$args --system-prompt-file $_ko_sp"
fi

# Max turns: default to 1 for single-shot agent work unless overridden
if [ -n "$KO_MAX_TURNS" ]; then
  args="$args --max-turns $KO_MAX_TURNS"
fi

# Pass prompt as positional argument
exec muse $args "$KO_PROMPT"
