#!/bin/sh
# Built-in harness for Claude Code CLI
# Receives parameters via KO_-namespaced environment variables

set -e

# Build command arguments
args="-p --output-format text"

# Add conditional flags only if set
if [ -n "$KO_ALLOW_ALL" ] && [ "$KO_ALLOW_ALL" = "true" ]; then
  args="$args --dangerously-skip-permissions"
fi

if [ -n "$KO_ALLOWED_TOOLS" ]; then
  args="$args --allowed-prompts $KO_ALLOWED_TOOLS"
fi

if [ -n "$KO_MODEL" ]; then
  args="$args --model $KO_MODEL"
fi

if [ -n "$KO_SYSTEM_PROMPT" ]; then
  args="$args --append-system-prompt $KO_SYSTEM_PROMPT"
fi

# Pass prompt via stdin
echo "$KO_PROMPT" | claude $args
