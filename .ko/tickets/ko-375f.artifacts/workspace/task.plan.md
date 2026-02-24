Perfect! I've written a comprehensive plan for migrating the agent harnesses from YAML templates to shell scripts. The plan includes:

## Summary

**Goal**: Replace the YAML-based agent harness template system with executable shell scripts for greater flexibility and extensibility.

**Context**: The current system uses YAML files with template variables (like `${prompt}`, `${model}`) that get rendered by Go code. This requires complex template logic in Go and makes it hard to add new features. Shell scripts would be more flexible and allow users to extend harnesses without modifying Go code.

**Approach**: Replace YAML templates with shell scripts that receive inputs via environment variables and output command-lines to stdout. The Go code executes the script and parses the output to build commands.

**Key Tasks**:
1. Replace YAML Harness/TemplateAdapter with shell script execution
2. Update LoadHarness to look for .sh files instead of .yaml
3. Convert claude.yaml → claude.sh
4. Convert cursor.yaml → cursor.sh  
5. Update embed directive for .sh files
6. Update all tests for shell-based approach
7. Remove YAML-specific tests
8. Update README documentation

**Open Questions** that need your input:
1. **Script output format**: Should scripts output full command lines or just args? How should stdin be signaled?
2. **Binary resolution**: Should scripts handle their own binary fallback logic?
3. **Error handling**: How should malformed script output be handled?
4. **Backward compatibility**: Hard cutover or support both YAML and shell during transition?

I've provided recommendations for each question, but these are architectural decisions that could benefit from your input before implementation begins.
