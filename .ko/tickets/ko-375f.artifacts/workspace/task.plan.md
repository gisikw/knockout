Perfect! I've created a comprehensive plan for migrating agent harnesses from YAML templates to shell scripts. The plan includes:

1. **Goal**: Clear one-sentence summary of the migration objective
2. **Context**: Detailed analysis of the current YAML template system and why shell scripts are superior
3. **Approach**: High-level strategy for the migration
4. **Tasks**: 10 detailed, ordered tasks covering all aspects of the migration from converting the harness files themselves to updating tests, documentation, and specs
5. **Open Questions**: 4 substantive architectural questions that need human input:
   - Backward compatibility strategy (breaking change vs. dual support)
   - Error handling behavior for script failures
   - Security considerations for executable scripts
   - Script permission handling

Each task includes specific file references, describes what needs to change and why, and includes concrete verification steps. The plan reflects the existing codebase patterns I found during investigation.
