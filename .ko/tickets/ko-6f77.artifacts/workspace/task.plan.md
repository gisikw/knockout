The plan is complete. I've investigated the codebase and identified:

1. **Current state**: `resolveProjectTicketsDir()` parses `#hashtags` from args for routing, and `RouteTicket()` parses hashtags from ticket titles in `ko add`

2. **Change required**: Replace both hashtag-based mechanisms with explicit `--project` flags

3. **Key files**:
   - cmd_list.go (core routing function + two commands)
   - cmd_loop.go (agent loop)
   - cmd_agent.go (agent start/stop/status)
   - cmd_build_init.go (agent init)
   - cmd_create.go (ticket creation)
   - registry.go (RouteTicket to be deprecated)

4. **Testing**: Following project patterns, tests should be added to verify flag behavior

The plan breaks down the work into 11 sequential tasks, each with clear verification steps.
