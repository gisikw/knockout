```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "Should auto-triage failure prevent the ticket from being created or updated?",
      "context": "When `ko add --triage` or `ko update --triage` triggers auto-triage and the triage operation fails (e.g., agent error), the plan currently treats this as non-fatal: the ticket saves with the triage field intact, a warning is printed, and the user can manually run `ko agent triage` afterward. The alternative is to fail the entire operation.",
      "options": [
        {
          "label": "Non-fatal (Recommended)",
          "value": "nonfatal",
          "description": "Ticket saves with triage field set; user sees warning and can retry manually. Safer because the ticket isn't lost."
        },
        {
          "label": "Fatal failure",
          "value": "fatal",
          "description": "If auto-triage fails, the entire create/update operation fails with error, preventing the ticket from being saved."
        }
      ]
    }
  ]
}
```
