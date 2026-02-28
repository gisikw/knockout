```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "Should `ko agent triage` enforce `allowAll=true` for tool calls, or respect the pipeline configuration?",
      "context": "Triage operations typically require running `ko` shell commands. The plan currently uses `p.AllowAll` from pipeline config, but this may restrict access on restrictive pipelines. Forcing `true` is more ergonomic but bypasses user permission settings.",
      "options": [
        {
          "label": "Force allowAll=true (Recommended)",
          "value": "force_allow_all",
          "description": "Always enable all tool calls for triage operations, regardless of pipeline config—more reliable for triage work"
        },
        {
          "label": "Use pipeline config setting",
          "value": "respect_config",
          "description": "Respect the pipeline's allow_all_tool_calls setting; users must configure it explicitly for triage to work"
        }
      ]
    },
    {
      "id": "q2",
      "question": "Should `ko agent triage` require a pipeline configuration to exist, or fall back to a default adapter?",
      "context": "The plan currently requires a pipeline config via `FindPipelineConfig`. The ticket says 'load default model', but it's unclear whether a pipeline is mandatory or if a hardcoded default (e.g., 'claude') should be used when no config exists.",
      "options": [
        {
          "label": "Require pipeline config",
          "value": "require_pipeline",
          "description": "Fail with a clear error if no pipeline exists—ensures intentional setup"
        },
        {
          "label": "Fall back to default (claude) (Recommended)",
          "value": "default_fallback",
          "description": "Use a hardcoded default adapter and model if no pipeline config exists—works on fresh projects"
        }
      ]
    }
  ]
}
```
