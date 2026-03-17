---
date: 2026-03-17
topic: milestone-a-usable-core
stage: implementation-plan
status: draft
source: conversation
---

# Milestone A Usable Core Implementation Plan

## Goal

Deliver a working Milestone A for `envdiff`: a deterministic, offline-first CLI with `compare`, `scan`, and `doctor` commands for `.env`, Python env access, and Docker Compose interpolation, with repo-local resolution and stable JSON output.

## Scope

### In Scope

- Python project bootstrap with `uv`, Typer CLI entrypoint, test harness, and lint/format setup.
- Core models for definitions, usages, contracts, findings, and JSON envelopes.
- Dotenv parsing with duplicates, comments, blank values, quoted values, warnings, and line tracking.
- Python AST and Docker Compose source scanners for the supported v1 syntax.
- Repo scanner and deterministic repo-local resolution for `.env` and `.env.example`.
- `compare`, `scan`, and `doctor` commands with human and JSON renderers.
- Structured findings for missing, unused, skew, duplicate, and base contract issues.
- Monorepo-aware behavior via nearest-definition resolution and documented fallback rules.

### Out of Scope

- Secret heuristics and alias detection beyond placeholder data-model support.
- `.envrc`, shell startup files, GitHub Actions, Pydantic settings, and other deferred parsers.
- Matrix mode, `.env.example` generation, baseline workflows, and plugin architecture.
- Runtime env loading, encryption, external secret managers, or automatic file rewrites.

## Assumptions And Constraints

- Python 3.11+ with `uv` is the project toolchain.
- Existing files `docs/project/SPEC.md`, `docs/project/ROADMAP.md`, `docs/research/MARKET_RESEARCH.md`, and the root `README.md` remain reference docs during implementation.
- Milestone A follows the updated roadmap: repo-local resolution is part of the core, not later polish.
- Stable JSON is a first-class product surface and should be modeled before command wiring.
- Determinism is mandatory: fixed ordering, fixed thresholds where used, no network access, no LLM dependencies.
- Heuristic-heavy work for aliases and secrets is intentionally deferred; the plan should not create premature abstractions for them.

## Task Breakdown

### Task 1: Bootstrap Project Skeleton

**Objective**

Create the Python package, CLI entrypoint, development tooling, and baseline test harness so all later tasks land on a stable project structure.

**Files**

- Create: `pyproject.toml`
- Create: `README.md`
- Create: `envdiff/__init__.py`
- Create: `envdiff/cli.py`
- Create: `envdiff/models.py`
- Create: `envdiff/parsers/__init__.py`
- Create: `envdiff/analyzers/__init__.py`
- Create: `envdiff/render/__init__.py`
- Create: `envdiff/utils/__init__.py`
- Create: `tests/__init__.py`
- Create: `tests/conftest.py`
- Create: `tests/fixtures/.gitkeep`

**Dependencies**

None

**Implementation Steps**

1. Initialize the `uv` project metadata and declare runtime and dev dependencies.
2. Add Typer CLI entrypoint scaffolding with placeholder subcommands returning non-zero or “not implemented” messages where needed.
3. Add lint and format configuration plus a base pytest configuration.
4. Add package directories and import-safe module stubs that subsequent tasks can extend.
5. Add a canonical `README.md` pointing back to the current design docs and stating the Milestone A scope.

**Verification**

- Run: `uv run python -m envdiff.cli --help`
- Expect: process exits `0` and prints the top-level CLI help.
- Run: `uv run pytest -q`
- Expect: test discovery completes successfully, even if only smoke tests exist.
- Run: `uv run ruff check .`
- Expect: exits `0`.

**Done When**

- The repo has a runnable CLI entrypoint.
- Package imports are clean.
- Test and lint commands are wired and passing in the bootstrap scope.

### Task 2: Implement Core Models And JSON Envelope

**Objective**

Define the shared data structures and deterministic serialization rules that all parsers, analyzers, and renderers will use.

**Files**

- Modify: `envdiff/models.py`
- Create: `envdiff/render/json.py`
- Create: `envdiff/utils/ordering.py`
- Create: `tests/test_models.py`

**Dependencies**

- Task 1

**Implementation Steps**

1. Add typed models for `EnvVarDefinition`, `EnvVarUsage`, `EnvVarContract`, `Finding`, command metadata, and top-level JSON result envelopes.
2. Encode stable ordering rules for findings, locations, and contract collections.
3. Add explicit schema-version support to the JSON envelope.
4. Add serialization helpers that guarantee machine-readable field names and deterministic output ordering.
5. Cover core model invariants with focused tests.

**Verification**

- Run: `uv run pytest tests/test_models.py -q`
- Expect: model tests pass.
- Run: `uv run python - <<'PY'\nfrom envdiff.models import JsonEnvelope\nprint(JsonEnvelope.schema_version())\nPY`
- Expect: prints a schema version string without import failures.

**Done When**

- All downstream tasks can depend on stable shared models.
- JSON output shape is defined before command implementations begin.
- Ordering rules are encoded in code and test-covered.

### Task 3: Build The Dotenv Parser

**Objective**

Implement robust dotenv parsing with warning capture and full source-location tracking.

**Files**

- Create: `envdiff/parsers/dotenv.py`
- Create: `tests/fixtures/dotenv/basic.env`
- Create: `tests/fixtures/dotenv/duplicates.env`
- Create: `tests/fixtures/dotenv/malformed.env`
- Create: `tests/test_dotenv.py`

**Dependencies**

- Task 2

**Implementation Steps**

1. Implement line-by-line parsing for active definitions, comments, quoted values, blanks, duplicates, and malformed lines.
2. Preserve source metadata and parse warnings for each definition.
3. Add normalized value-kind helpers only to the extent required for compare and doctor base checks.
4. Create fixture files for positive and negative parsing paths.
5. Add deterministic parser tests, including ordering and duplicate preservation.

**Verification**

- Run: `uv run pytest tests/test_dotenv.py -q`
- Expect: dotenv parser tests pass.
- Run: `uv run python - <<'PY'\nfrom envdiff.parsers.dotenv import parse_dotenv\nresult = parse_dotenv('tests/fixtures/dotenv/duplicates.env')\nprint(len(result.definitions), len(result.warnings))\nPY`
- Expect: prints stable counts and does not crash.

**Done When**

- Dotenv parsing supports all documented v1 syntax.
- Duplicate definitions are preserved, not collapsed away.
- Warnings and line numbers are available to analyzers.

### Task 4: Build Python And Compose Usage Scanners

**Objective**

Infer required and optional env-variable usage from Python and Docker Compose sources with soft-fail behavior on unsupported constructs.

**Files**

- Create: `envdiff/parsers/python_ast.py`
- Create: `envdiff/parsers/compose.py`
- Create: `tests/fixtures/python/sample_app.py`
- Create: `tests/fixtures/python/unsupported.py`
- Create: `tests/fixtures/compose/docker-compose.yml`
- Create: `tests/test_python_ast.py`
- Create: `tests/test_compose.py`

**Dependencies**

- Task 2

**Implementation Steps**

1. Implement Python AST detection for `os.environ["X"]`, `os.getenv("X")`, and `os.getenv("X", default)`.
2. Implement Docker Compose interpolation scanning for `${VAR}` and `${VAR:-default}`.
3. Emit `EnvVarUsage` records with file path, line number, requiredness, and default where applicable.
4. Ensure unsupported patterns are ignored or warned on without crashing.
5. Add fixture-driven tests for required, optional, defaulted, and unsupported cases.

**Verification**

- Run: `uv run pytest tests/test_python_ast.py tests/test_compose.py -q`
- Expect: both scanner suites pass.
- Run: `uv run python - <<'PY'\nfrom envdiff.parsers.python_ast import scan_python_file\nprint(len(scan_python_file('tests/fixtures/python/sample_app.py').usages))\nPY`
- Expect: prints a stable usage count.

**Done When**

- Supported Python and Compose patterns produce correct usage classifications.
- File and line metadata are present.
- Unsupported syntax fails soft.

### Task 5: Implement Repo Scan And Resolution

**Objective**

Aggregate definitions and usages into contracts and make repo-local association deterministic for monorepos and multi-service layouts.

**Files**

- Create: `envdiff/analyzers/scan.py`
- Create: `envdiff/utils/paths.py`
- Create: `tests/fixtures/repos/simple_repo/...`
- Create: `tests/fixtures/repos/monorepo/...`
- Create: `tests/test_scan.py`

**Dependencies**

- Task 3
- Task 4

**Implementation Steps**

1. Implement deterministic repo traversal with ignore support.
2. Define and codify nearest-definition / nearest-example association rules.
3. Aggregate parser outputs into `EnvVarContract` records with statuses such as referenced, defined, required, optional, and undefined.
4. Record provenance showing which env files were associated with which usage sets.
5. Add repo fixtures covering single-service and monorepo layouts, including ambiguous or fallback cases.

**Verification**

- Run: `uv run pytest tests/test_scan.py -q`
- Expect: scan and resolution tests pass.
- Run: `uv run python -m envdiff.cli scan tests/fixtures/repos/monorepo --json`
- Expect: exits `0` and emits stable JSON with contracts, locations, and resolution metadata.

**Done When**

- Repo scans are deterministic.
- Monorepo association behavior is documented in tests.
- Scan output is sufficient input for both `compare` and `doctor`.

### Task 6: Implement Compare Command

**Objective**

Ship the first fully user-visible command for file-to-file dotenv comparison in both human and JSON modes.

**Files**

- Create: `envdiff/analyzers/compare.py`
- Create: `envdiff/render/human.py`
- Modify: `envdiff/cli.py`
- Create: `tests/fixtures/compare/left.env`
- Create: `tests/fixtures/compare/right.env`
- Create: `tests/test_compare.py`

**Dependencies**

- Task 3
- Task 2

**Implementation Steps**

1. Implement compare logic for missing, extra, duplicate, and value-kind differences.
2. Add deterministic ordering and summary counts for compare results.
3. Implement human-rendered output sections and JSON rendering through the shared envelope.
4. Wire the `compare` command into the CLI with `--json`.
5. Add focused tests for both analyzer logic and command-level output shape.

**Verification**

- Run: `uv run pytest tests/test_compare.py -q`
- Expect: compare tests pass.
- Run: `uv run python -m envdiff.cli compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env --json`
- Expect: exits `0` and emits machine-readable compare results.

**Done When**

- `compare` works end to end in human and JSON modes.
- Value-kind comparison is deterministic.
- Output includes duplicates and asymmetric keys correctly.

### Task 7: Implement Doctor Command

**Objective**

Build repo-level validation on top of the scan contract model and emit structured findings suitable for terminal use and CI.

**Files**

- Create: `envdiff/analyzers/doctor.py`
- Modify: `envdiff/cli.py`
- Create: `tests/fixtures/doctor/project/...`
- Create: `tests/test_doctor.py`

**Dependencies**

- Task 5
- Task 2

**Implementation Steps**

1. Implement finding generation for missing required, missing optional, defined-but-unused, duplicates, and `.env` vs `.env.example` skew.
2. Assign stable finding codes, severities, reasons, and machine-readable locations.
3. Implement `--fail-on` threshold handling with exit code `2` for findings at or above the threshold.
4. Render doctor findings in both human and JSON modes using the shared envelope.
5. Add tests for positive, negative, and threshold-driven exit behavior.

**Verification**

- Run: `uv run pytest tests/test_doctor.py -q`
- Expect: doctor tests pass.
- Run: `uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --fail-on warning`
- Expect: exits `2` when warning-or-higher findings are present.
- Run: `uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --json`
- Expect: emits stable findings with codes, severities, and locations.

**Done When**

- `doctor` provides actionable structured findings.
- Exit code semantics match the spec.
- JSON output is stable enough for CI and agent consumption.

### Task 8: Milestone A Integration Pass

**Objective**

Finish command wiring, regression coverage, and minimal user-facing documentation so the core tool is ready for implementation follow-through.

**Files**

- Modify: `envdiff/cli.py`
- Modify: `README.md`
- Modify: `pyproject.toml`
- Create: `tests/test_cli_smoke.py`

**Dependencies**

- Task 6
- Task 7

**Implementation Steps**

1. Ensure all major commands share common option behavior and error handling.
2. Add smoke tests for `--help`, `compare`, `scan`, and `doctor`.
3. Update `README.md` with current supported syntax and command examples.
4. Verify lint, tests, and command-level smoke checks as a Milestone A gate.

**Verification**

- Run: `uv run pytest -q`
- Expect: full test suite passes.
- Run: `uv run ruff check .`
- Expect: exits `0`.
- Run: `uv run python -m envdiff.cli --help`
- Expect: exits `0`.
- Run: `uv run python -m envdiff.cli scan tests/fixtures/repos/simple_repo --json`
- Expect: exits `0` and emits stable JSON.

**Done When**

- Milestone A commands are wired consistently.
- The supported syntax and scope are documented.
- A full local verification pass succeeds.

## Risks And Mitigations

- Risk: repo-local resolution rules become ambiguous in monorepos.
  Mitigation: codify a narrow nearest-definition algorithm, capture provenance in output, and lock behavior with fixture tests before expanding parser coverage.
- Risk: JSON shape drifts while analyzers are still evolving.
  Mitigation: define the envelope and schema version in Task 2 and require command tests to assert explicit fields.
- Risk: dotenv parser edge cases consume too much time early.
  Mitigation: implement only the v1 syntax from the spec, preserve warnings instead of chasing shell semantics, and defer non-spec syntax.
- Risk: doctor output becomes a dumping ground for future heuristics.
  Mitigation: keep Task 7 limited to deterministic core findings and leave alias/secret heuristics for the next milestone.
- Risk: premature extensibility slows delivery.
  Mitigation: avoid plugin abstractions until after Milestone A proves the model and command surface.

## Verification Matrix

| Requirement | Proof command | Expected signal |
| --- | --- | --- |
| CLI bootstrap works | `uv run python -m envdiff.cli --help` | Exit `0` and visible subcommand help |
| Shared models are importable and stable | `uv run pytest tests/test_models.py -q` | Model tests pass |
| Dotenv parsing covers v1 syntax | `uv run pytest tests/test_dotenv.py -q` | Parser tests pass |
| Python and Compose scanning infer requiredness | `uv run pytest tests/test_python_ast.py tests/test_compose.py -q` | Scanner tests pass |
| Repo scan is deterministic and monorepo-aware | `uv run pytest tests/test_scan.py -q` | Scan tests pass |
| Compare works in JSON mode | `uv run python -m envdiff.cli compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env --json` | Exit `0` and stable JSON output |
| Doctor threshold behavior matches spec | `uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --fail-on warning` | Exit `2` when warning-or-higher findings exist |
| Full Milestone A passes local quality gates | `uv run pytest -q && uv run ruff check .` | All tests pass and lint exits `0` |

## Handoff

Parallel execution notes:

- After Task 2 lands, Task 3 and Task 4 can run in parallel because they share only the model layer.
- After Task 3 lands, Task 6 can start in parallel with Task 5, but only for analyzer work that depends on dotenv definitions rather than full repo aggregation.
- Task 7 should wait for Task 5 because `doctor` depends on the repo contract model and resolution behavior.
- Task 8 is the convergence gate and should remain sequential.

Recommended execution order:

1. Task 1
2. Task 2
3. Task 3 and Task 4 in parallel
4. Task 5 and Task 6 with partial overlap
5. Task 7
6. Task 8

Plan complete and saved to `docs/plans/2026-03-17-milestone-a-usable-core-implementation.md`.

1. Execute in this session, task by task.
2. Open a separate execution session.
3. Refine this plan before implementation.
