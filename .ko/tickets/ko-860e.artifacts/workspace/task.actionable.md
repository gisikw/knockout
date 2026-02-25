The plan has **three open questions** that require human input before implementation can proceed. Let me convert these into structured plan questions:

```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "What format should bare `ko triage <id>` use to display block reason and open questions?",
      "context": "The ticket spec says 'show block reason + open questions' but doesn't specify format. The current assumption is human-readable text with block reason followed by JSON array of questions, matching existing `ko questions` output.",
      "options": [
        {
          "label": "Human-readable with JSON (Recommended)",
          "value": "human_json_hybrid",
          "description": "Show block reason as text, followed by JSON array of questions for machine-readability"
        },
        {
          "label": "JSON-only",
          "value": "json_only",
          "description": "Output a single JSON object with block reason and questions fields"
        },
        {
          "label": "Human-readable only",
          "value": "human_only",
          "description": "Format block reason and questions as human-readable text without JSON"
        }
      ]
    },
    {
      "id": "q2",
      "question": "What note format should `ko triage <id> --block <reason>` use when saving the block reason?",
      "context": "The existing `ExtractBlockReason` function parses notes in the format 'ko: BLOCKED — {reason}'. The plan assumes this format should be used for consistency, but we need to confirm.",
      "options": [
        {
          "label": "Parseable format (Recommended)",
          "value": "parseable_format",
          "description": "Use 'ko: BLOCKED — {reason}' format so ExtractBlockReason can extract it consistently"
        },
        {
          "label": "Plaintext format",
          "value": "plaintext_format",
          "description": "Store reason as plain text without the 'ko: BLOCKED' prefix"
        }
      ]
    },
    {
      "id": "q3",
      "question": "Should this be a breaking change by removing the old commands entirely, or maintain backward compatibility?",
      "context": "The ticket says 'Replaces' which implies removal. However, this breaks existing scripts that call `ko block`, `ko questions`, or `ko answer`. The plan assumes full removal per ticket spec.",
      "options": [
        {
          "label": "Remove entirely (Recommended)",
          "value": "breaking_change",
          "description": "Delete old commands as specified in ticket, breaking existing scripts"
        },
        {
          "label": "Keep as aliases",
          "value": "aliases",
          "description": "Maintain old commands as wrappers that call `ko triage` for backward compatibility"
        }
      ]
    }
  ]
}
```
