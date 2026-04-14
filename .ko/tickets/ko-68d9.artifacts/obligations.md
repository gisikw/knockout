## Obligations

1. [observable] `~/.config/knockout/config.yaml` contains a top-level key (e.g., `projects`) that holds the project registry previously stored in `projects.yml`.
   Check: Run `cat ~/.config/knockout/config.yaml` and confirm a projects section is present with the same entries that were in `projects.yml`.

2. [observable] `~/.config/knockout/projects.yml` no longer exists (or is ignored) after migration.
   Check: Run `ls ~/.config/knockout/` and confirm `projects.yml` is absent, or confirm knockout does not read it.

3. [observable] The knockout CLI reads project registry data from `config.yaml` only — listing/using projects works without `projects.yml` present.
   Check: Remove `projects.yml` (or rename it), run a project-listing command (e.g., `knockout list` or equivalent), and confirm projects are still returned correctly.

4. [preserved] All project entries from the original `projects.yml` are accessible after consolidation — no projects are lost or renamed.
   Check: Compare the project list before and after migration; all previously registered projects must still appear.

5. [preserved] Existing settings in `config.yaml` (non-project configuration) are unmodified and still take effect.
   Check: Inspect a known setting (e.g., a default editor, output format, or any documented option) and confirm it behaves identically before and after.
