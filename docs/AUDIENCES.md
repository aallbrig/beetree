# BeeTree Target Audiences

## Audience 1: Beginner Game Developer / Designer

**Who they are:** Students, hobbyists, and designers new to behavior trees.

**What they need from BeeTree:**
- A short, obvious first-run path from install → first tree → simulation → engine integration
- Plain-language BT concepts in-tool and in docs
- Progressive examples that map directly to real game AI scenarios
- Safety rails in editor workflows (validation, warnings, clear recovery guidance)

**Current friction:**
- Discovery and onboarding are still doc-light outside the README/examples
- Error language still assumes BT vocabulary in several paths

**Success metric:** A beginner can reach generated, integrated AI behavior in under 30 minutes without leaving official BeeTree docs.

---

## Audience 2: Intermediate Game Developer

**Who they are:** Working developers comfortable with YAML/CLI and BT basics.

**What they need from BeeTree:**
- Fast authoring/editing loop (CLI + TUI)
- Reliable validation, simulation, rendering, and diffing for day-to-day workflow
- Predictable regeneration behavior for generated code and stubs
- Strong docs for structure, conventions, and migration practices

**Current friction:**
- TUI edits many core fields, but advanced spec areas still require manual YAML edits
- Registry remains hidden/stubbed, limiting multi-user workflows

**Success metric:** An intermediate team member can refactor and ship multiple tree specs per sprint with low manual rework.

---

## Audience 3: Expert / AI Architect

**Who they are:** Senior engineers or AI leads evaluating BeeTree as a team standard.

**What they need from BeeTree:**
- Spec expressiveness and stability guarantees
- Team workflows (reviewability, CI usage, registry/distribution story)
- Extensibility points for validation/generation/runtime conventions
- Operational docs that reduce ambiguity across contributors

**Current friction:**
- No real registry backend or explicit extensibility model
- No formal SDD document set that defines architecture boundaries, evolution rules, and compatibility policies

**Success metric:** An expert can standardize BeeTree-based AI authoring across a team with clear governance and low integration risk.

---

## Audience 4: Agent Session Contributor (AI-assisted development)

**Who they are:** Future Copilot/Claude agent sessions implementing incremental features.

**What they need from BeeTree docs:**
- Explicit specs of intended behavior, invariants, and acceptance criteria per subsystem
- Stable document index describing where architecture decisions and contracts live
- Guidance for safe, minimal changes without rediscovering domain context

**Current friction:**
- Core behavior exists in code/tests, but intent and boundaries are not consolidated in spec-driven documents

**Success metric:** A new agent session can deliver a correct feature with minimal exploratory churn and no semantic regressions.

---

## Cross-Cutting Priority Matrix

| Need | Beginner | Intermediate | Expert | Agent Contributor |
|------|:--------:|:------------:|:------:|:-----------------:|
| Clear onboarding docs | Critical | High | Medium | Medium |
| Example quality | Critical | High | Medium | Medium |
| TUI completeness | High | Critical | Medium | Medium |
| CLI scriptability | Medium | High | Critical | High |
| Spec/contract documentation (SDD) | Medium | High | Critical | Critical |
| Code generation guidance | High | Critical | Critical | High |
| Error message quality | Critical | High | Medium | High |
| CI-friendly validation | Medium | High | Critical | Critical |
| Registry/team sharing | Low | Medium | Critical | Medium |
