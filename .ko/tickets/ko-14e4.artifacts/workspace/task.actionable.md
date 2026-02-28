The plan has **three open questions** that require human input before implementation can proceed. I'll convert these to structured questions:

```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "How should the loop handle triage failures?",
      "context": "If a ticket's triage fails (agent exits non-zero, timeout, etc.), the loop could either log the error and continue to the next ticket, or stop the loop entirely with a build_error status.",
      "options": [
        {
          "label": "Log and continue (Recommended)",
          "value": "log_continue",
          "description": "Log triage failures but continue processing remaining tickets, consistent with how FAIL outcomes work"
        },
        {
          "label": "Stop the loop",
          "value": "stop_on_failure",
          "description": "Halt the loop when triage fails, marking the run with a build_error status"
        }
      ]
    },
    {
      "id": "q2",
      "question": "Should triaged tickets count toward the --max-tickets limit?",
      "context": "The plan treats triage as a pre-build housekeeping step. Deciding whether triage counts toward --max-tickets affects how many ready tickets can be processed in a single loop run.",
      "options": [
        {
          "label": "Do not count (Recommended)",
          "value": "no_count",
          "description": "Triage is housekeeping; only ready-to-build tickets consume the limit"
        },
        {
          "label": "Count toward limit",
          "value": "count_limit",
          "description": "Triaged tickets count against --max-tickets, reducing the number of builds possible"
        }
      ]
    },
    {
      "id": "q3",
      "question": "When should the loop honor a stop signal during triage?",
      "context": "A stop signal could be checked between each triage run, or it could interrupt an in-progress triage immediately. The plan assumes inter-ticket gaps, consistent with how stop signals are checked between builds today.",
      "options": [
        {
          "label": "Check between tickets (Recommended)",
          "value": "inter_ticket",
          "description": "Honor stop signals only between triage runs, allowing current triage to complete"
        },
        {
          "label": "Interrupt immediately",
          "value": "immediate",
          "description": "Stop the current triage run immediately when a signal arrives"
        }
      ]
    }
  ]
}
```
