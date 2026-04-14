## Obligations

1. [observable] A pipeline ticket (or pipeline step) can be annotated or configured to run in parallel mode, distinct from the default sequential mode.
   Check: Inspect the pipeline configuration format (e.g. YAML/JSON/config file) and confirm there is a field or flag (e.g. `parallel: true`, `mode: parallel`) that can be set on a ticket or step.

2. [observable] When pipeline steps are marked as parallel, they are dispatched concurrently rather than waiting for each preceding step to complete.
   Check: Create a pipeline with two parallel research steps and observe (via logs or timing) that both begin execution before either finishes.

3. [observable] Parallel steps are not subject to the same sequential ordering restrictions as normal pipeline steps (e.g. they do not block on dependency completion in the same way).
   Check: Define two parallel steps with no declared dependency between them; confirm they start simultaneously rather than one waiting on the other.

4. [observable] Sequential (non-parallel) pipeline steps continue to execute in order and are unaffected by the new parallelism feature.
   Check: Run an existing pipeline that uses no parallel annotation and confirm steps execute one at a time in declared order, same as before.

5. [preserved] Existing pipelines without any parallelism configuration continue to run correctly with the same sequential behavior as before.
   Check: Run the existing test suite or a known pipeline definition without modification and confirm it produces the same output/order as prior to this change.
