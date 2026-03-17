# envdiff Roadmap

## Overview

This roadmap is designed for terminal-based implementation by AI agents. Each phase contains a clear objective, bounded deliverables, acceptance criteria, and explicit non-goals to limit scope drift.

The recommended development sequence is:

1. deterministic parsing
2. contract inference
3. comparison + linting
4. UX + CI usability
5. broader parser coverage
6. generation and multi-environment intelligence

All phases must preserve:
- deterministic output
- offline-first behavior
- stable JSON for agent consumption
- repo-local scope boundaries
- explainable heuristics when confidence-based findings are used
- low abstraction overhead until justified

---

## Phase 0 — Bootstrap and Project Skeleton

### Objective
Create the repo skeleton and core CLI entrypoint.

### Deliverables
- Python project initialized with `uv`
- `Typer` CLI entrypoint
- package structure for parsers, analyzers, renderers, and models
- base finding model and JSON envelope shape
- base test harness
- sample fixture files for `.env`, Python, and Docker Compose

### Acceptance Criteria
- `envdiff --help` works
- package imports cleanly
- finding / JSON envelope models import cleanly
- tests execute successfully
- lint/format tooling configured

### Non-goals
- no parser logic yet
- no analysis logic yet
- no plugin system yet

---

## Phase 1 — Dotenv Parsing Foundation

### Objective
Implement reliable dotenv parsing.

### Deliverables
- dotenv parser
- support for comments, blank values, duplicates, quoted values
- parse warnings model
- normalized definition objects
- tests for malformed and edge-case inputs

### Acceptance Criteria
- parser handles common dotenv syntax without crashing
- duplicate keys are preserved and flagged
- line numbers are recorded
- deterministic output ordering

### Non-goals
- no code scanning yet
- no alias detection yet

---

## Phase 2 — Source Usage Scanners

### Objective
Infer expected env vars from source/config files.

### Deliverables
- Python AST scanner
  - `os.environ["X"]`
  - `os.getenv("X")`
  - `os.getenv("X", default)`
- Docker Compose scanner
  - `${VAR}`
  - `${VAR:-default}`
- unified usage model
- fixture-based tests

### Acceptance Criteria
- required vs optional inference works for supported patterns
- file path and line number captured
- unsupported constructs fail soft rather than crashing

### Non-goals
- no shell parsing
- no Pydantic inference yet
- no GitHub Actions support yet

---

## Phase 3 — Contract Model and Repo Scan

### Objective
Build the internal contract model for combining usages and definitions.

### Deliverables
- `EnvVarContract` aggregation layer
- repo scanner for collecting definitions and usages
- repo-local resolution model for associating usages with `.env` / `.env.example`
- nearest-definition behavior for monorepos and multi-service repos
- ignore-path support
- deterministic traversal order
- basic summary output for scan mode

### Acceptance Criteria
- variables can be classified as referenced/defined/required/optional
- locations are preserved
- repo-local resolution behavior is deterministic and documented
- repo scan is deterministic
- scan results usable by analyzers

### Non-goals
- no compare command yet
- no doctor command yet

---

## Phase 4 — Compare Command

### Objective
Implement file-to-file comparison.

### Deliverables
- `envdiff compare <left> <right>`
- detection for missing / extra / duplicate / differing value-class
- human renderer
- JSON renderer
- tests for compare scenarios

### Acceptance Criteria
- compare works for two dotenv files
- JSON output stable and machine-readable
- value normalization categories are applied consistently

### Non-goals
- no multi-file matrix yet
- no alias detection yet

---

## Phase 5 — Doctor Command

### Objective
Implement repo-level contract validation.

### Deliverables
- `envdiff doctor <path>`
- finding codes and severities
- confidence and reason fields for heuristic-ready findings
- missing required/optional detection
- defined-but-unused detection
- environment skew checks against `.env.example` where present
- `--fail-on` exit behavior

### Acceptance Criteria
- doctor produces actionable findings
- exit code behavior matches severity threshold
- findings contain locations and suggested fixes where possible

### Non-goals
- no secret heuristics yet
- no alias detection yet

---

## Phase 6 — Alias and Secret Heuristics

### Objective
Add the two highest-value heuristic layers: naming drift and secret hygiene.

### Deliverables
- alias similarity engine
- fixed-threshold candidate scoring
- secret-like value detector
- placeholder detector
- explainable reasons for alias/secret findings
- conservative thresholds tuned to minimize noisy output
- doctor integration for both

### Acceptance Criteria
- common alias cases are detected with explainable heuristics
- placeholders are not misclassified as secrets
- ordering and scores remain deterministic

### Non-goals
- no network calls
- no external secret APIs

---

## Phase 7 — Output Quality and CI Usability

### Objective
Make the tool practical for repeated terminal and CI use.

### Deliverables
- improved Rich rendering
- stable JSON contract documentation
- schema versioning for machine-readable output
- `--quiet`, `--verbose`, and `--json` polish
- clearer exit codes
- summary counts per severity
- performance pass for medium-sized repos

### Acceptance Criteria
- human output readable in standard terminals
- JSON output documented and stable
- CI workflows can fail on warning/error predictably

### Non-goals
- no new parser categories

---

## Phase 8 — Generation and Matrix Modes

### Objective
Add broader contract workflows.

### Deliverables
- `envdiff matrix`
- `envdiff generate`
- annotated `.env.example` generation from inferred contract
- multi-env consistency summaries
- baseline / snapshot workflow for incremental adoption if prior phases remain bounded

### Acceptance Criteria
- matrix compares multiple env files deterministically
- generate emits valid dotenv output
- generated file can include comments/annotations when enabled
- baseline mode, if included, is deterministic and supports gradual adoption in CI

### Non-goals
- no automatic repo rewriting
- no cloud/env injection support

---

## Phase 9 — Expanded Parser Coverage

### Objective
Broaden ecosystem support after core value is proven.

### Candidate Deliverables
- GitHub Actions parser
- shell script parser
- repo-local `.envrc` / direnv-style parser using a conservative supported subset
- explicit non-goal: user shell startup files such as `~/.zshrc` and `~/.zprofile`
- Pydantic/BaseSettings inference
- `.devcontainer/devcontainer.json` support
- service grouping for monorepos

### Acceptance Criteria
- each new parser has fixture coverage
- unsupported syntax degrades gracefully
- no regressions in base commands

---

## Global Engineering Rules for Agents

### Determinism
Same repo and same flags must produce identical outputs.

### Offline-first
Do not depend on network access for core functionality.

### Stable JSON
All machine-readable outputs must have explicit, documented fields.

### Product Boundary
Do not drift into env loading, encryption/distribution, or machine-global shell configuration.

### Fail Soft
Malformed inputs should produce warnings where possible, not hard crashes.

### Bounded Scope
Do not implement speculative integrations before the prior phase is complete.

### Test Before Extension
Add fixtures and tests before broadening parser support.

### Defer Plugin Systems
Do not introduce a general plugin architecture before the core analyzers and JSON contract are proven.

---

## Suggested Agent Execution Order

1. bootstrap CLI and tests
2. dotenv parser
3. Python and Compose scanners
4. contract aggregation and repo-local resolution
5. compare command
6. doctor command and structured findings
7. alias and secret heuristics
8. output/CI polish
9. matrix, generate, and optional baseline workflow
10. ecosystem parser expansion

---

## Recommended Milestones

### Milestone A — Usable Core
Phases 0–5 complete  
Outcome: compare, scan, and doctor are real, useful, and monorepo-aware.

### Milestone B — Practical Team Tool
Phase 6–7 complete  
Outcome: alias detection, secret hygiene, CI friendliness, and stable machine contracts.

### Milestone C — Broader Contract Platform
Phase 8–9 complete  
Outcome: matrix, generation, and additional parser support.

---

## Stretch Ideas

- pre-commit integration
- editor diagnostics
- autofix suggestions for alias standardization
- config contract export/import
