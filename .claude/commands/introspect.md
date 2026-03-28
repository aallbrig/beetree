Run an introspect checkpoint on the beetree project.

An introspect checkpoint is a deep usability analysis of the current state of the software (CLI, TUI, website, code generation, examples) evaluated against the target audiences defined in `docs/AUDIENCES.md`.

## Steps

1. **Read current state**: Read all key source files, the existing `docs/AUDIENCES.md`, `docs/FEATURES.md`, and the most recent checkpoint in `docs/introspect/`. Run `make install && treemand beetree` to capture the current CLI command tree and include the output in the checkpoint.

2. **Evaluate from each audience perspective** (beginner / intermediate / expert game dev):
   - First-run experience: what happens when someone installs and runs `beetree`?
   - Command discoverability: are commands named intuitively? Is help clear?
   - TUI completeness: what can you do? What can't you do that you should?
   - Simulation UX: is it understandable for someone new to behavior trees?
   - Code generation path: is the workflow from tree → engine code clear?
   - Error messages: helpful or cryptic?
   - Example quality: do they teach BT concepts progressively?
   - Website/docs state: is documentation sufficient?

3. **Update docs**:
   - Update `docs/AUDIENCES.md` if audience understanding has evolved
   - Update `docs/FEATURES.md` with current feature status and gaps
   - Create `docs/introspect/$(date -u +%Y%m%d_%H%M%S)_UTC_CHECKPOINT.md` with:
     - Current state summary (version, components, test coverage)
     - CLI command tree (output of `treemand beetree` in a code block)
     - Alignment score per audience (1-10)
     - Killer feature health assessment
     - Gap analysis (critical / major / minor)
     - Prioritized improvement plan
     - Metrics for next checkpoint

4. **Compare to previous checkpoint**: Note what improved, what regressed, and what's unchanged since the last checkpoint.

5. **Be brutally honest.** The goal is to identify what sucks and why, not to be encouraging.

Commit the updated docs when complete.
