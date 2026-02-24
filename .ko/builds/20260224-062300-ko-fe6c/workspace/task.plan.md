Plan written. The plan addresses the ticket requirement to add multiline text support via heredoc/stdin while maintaining backward compatibility with the current command-line argument approach. The key decisions are:

1. **Stdin detection**: Use `os.Stdin.Stat()` to check if stdin is a pipe
2. **Precedence**: Stdin takes precedence over args when it's a pipe (standard Unix convention)
3. **Normalization**: Trim whitespace while preserving internal newlines
4. **Spec-first**: Add behavioral specs before implementation per INVARIANTS.md

The plan is proportional to the task and includes concrete verification steps for each task.
