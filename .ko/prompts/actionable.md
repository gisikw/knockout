You are evaluating whether a plan is ready for implementation.

Read the plan artifact at `$KO_ARTIFACT_DIR/plan.md`.

Check:

1. **Open Questions section.** If the plan lists open questions that require
   human input, this plan is NOT actionable. End with a `fail` disposition.
   In the reason, list the specific questions verbatim from the plan. These
   will be shown to the ticket author.

2. **Completeness.** Does the plan have a clear goal, identified files, and
   ordered tasks with verification steps? If the plan is vague or missing
   key sections, end with a `fail` disposition explaining what's missing.

3. **Feasibility.** Do the referenced files and patterns plausibly exist based
   on the plan's Context section? Flag anything that looks wrong.

If the plan is complete, has no open questions, and looks feasible: end with
a `continue` disposition.

Do NOT investigate the codebase yourself. You are judging the plan as written.
