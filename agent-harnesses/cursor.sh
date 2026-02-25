#!/bin/sh
# Built-in harness for Cursor agent CLI
# Cursor doesn't support separate system prompt, so we combine them

set -e

# Binary fallback: try cursor-agent first, then agent
if command -v cursor-agent >/dev/null 2>&1; then
  BINARY="cursor-agent"
elif command -v agent >/dev/null 2>&1; then
  BINARY="agent"
else
  # Fall back to cursor-agent (will fail with clear error message)
  BINARY="cursor-agent"
fi

# Combine system prompt and user prompt
if [ -n "$KO_SYSTEM_PROMPT" ]; then
  COMBINED_PROMPT="$KO_SYSTEM_PROMPT

$KO_PROMPT"
else
  COMBINED_PROMPT="$KO_PROMPT"
fi

# Build command arguments
args="--output-format text"

# Add conditional flags only if set
if [ -n "$KO_ALLOW_ALL" ] && [ "$KO_ALLOW_ALL" = "true" ]; then
  args="$args --force"
fi

if [ -n "$KO_ALLOWED_TOOLS" ]; then
  args="$args --allowed-prompts $KO_ALLOWED_TOOLS"
fi

if [ -n "$KO_MODEL" ]; then
  args="$args --model $KO_MODEL"
fi

# Execute the agent with prompt as separate argument (properly quoted)
exec $BINARY -p "$COMBINED_PROMPT" $args
