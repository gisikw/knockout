Perfect! I've created the implementation plan. Let me summarize what I found and planned:

## Summary

I've investigated the codebase and created a plan at `/home/dev/Projects/knockout/.ko/tickets/ko-f21c.artifacts/plan.md`.

**Key findings:**
- The pipeline parser in `pipeline.go` uses a minimal handwritten YAML parser that processes line-by-line
- It already handles multiline constructs for routes and hook commands
- The `prompt:` field currently stores just a filename string, which gets loaded via `LoadPromptFile()` 
- The parser needs to detect the YAML pipe syntax (`|`) and accumulate indented lines

**The plan implements inline prompt support in 4 tasks:**
1. Extend the YAML parser to handle `prompt: |` multiline syntax
2. Modify `runPromptNode()` to use inline content directly when it contains newlines
3. Add a test for inline prompts
4. Add a test for mixed inline/file-based prompts in the same workflow

The approach maintains backward compatibility by distinguishing inline prompts (contain newlines) from file references (no whitespace), which is unambiguous.
