knockout — just --list for recipes
**Obligations summary:**

1. **[observable]** A pipeline ticket/step can be marked for parallel execution via a config field or flag.
2. **[observable]** Parallel-marked steps are dispatched concurrently — both start before either finishes.
3. **[observable]** Parallel steps are not subject to sequential ordering restrictions (no blocking on preceding steps).
4. **[observable]** Non-parallel steps continue to execute in declared order, unaffected by the new feature.
5. **[preserved]** Existing pipelines without parallelism config run identically to before — no sequential behavior broken.
