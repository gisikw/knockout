knockout — just --list for recipes
**Obligations summary:**

1. **[observable]** `config.yaml` gains a projects section with all registry data from `projects.yml`.
2. **[observable]** `projects.yml` no longer exists or is no longer read.
3. **[observable]** Project listing/usage works with `projects.yml` absent (single source of truth).
4. **[preserved]** All previously registered projects survive the migration intact.
5. **[preserved]** Existing non-project settings in `config.yaml` continue to work unchanged.
