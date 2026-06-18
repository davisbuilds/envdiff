---
date: 2026-06-18
topic: go-port
stage: spec
status: complete
source: conversation
---

# Go Port Spec

## Goal

Port `envdiff` from Python to Go inside the same repository while preserving the
current CLI behavior, JSON schema version `1`, finding codes, deterministic
ordering, local-only analysis boundary, and fixture-backed test surface.

## Scope

### In Scope

- Add a side-by-side Go implementation under `cmd/envdiff/` and `internal/`.
- Keep the existing Python implementation as the release path and oracle until
  Go parity is proven.
- Preserve the documented command surface:
  - `compare <left> <right> [--json]`
  - `scan <path> [--json]`
  - `matrix <paths...> [--show-all] [--json]`
  - `generate <path> [--annotate] [--check] [--output <path>] [--json]`
  - `doctor <path> [--fail-on <severity>] [--baseline <path>]`
    `[--write-baseline <path>] [--ignore-file <path>] [--json]`
- Preserve exit codes:
  - `0` for success or no findings at the configured threshold.
  - `1` for command usage, parse, or execution failures.
  - `2` for findings or generated-template drift at the configured threshold.
- Preserve JSON envelope shape, field names, nullability, and deterministic
  ordering for schema version `1`.
- Preserve finding codes `ENV001` through `ENV009`, severities, suppression key
  formats, baseline snapshot format, and ignore-file behavior.
- Reuse the existing fixture repos and dotenv/source fixtures.
- Add Python-oracle golden coverage for command-level parity.
- Add Go unit tests for parser, analyzer, rendering, and CLI behavior.
- Update docs and CI only after the side-by-side Go implementation exists.
- Switch `./envdiff` to the Go implementation only in the final parity task.

### Out of Scope

- New parser categories or broader syntax support.
- Changes to JSON schema version `1`.
- Changes to finding-code meanings or severities.
- Any env loading, secret distribution, encryption, or shell startup file parsing.
- Network access in core analysis.
- A plugin system.
- Rewriting the Python implementation while the Go port is in progress.
- Removing Python before Go has matched the current test and parity gates.

## Assumptions And Constraints

- The active repo root is `/Users/dg-mac-mini/Dev/envdiff`.
- Go was installed during Task 1 with Homebrew; `go version` reports
  `go1.26.4 darwin/arm64`, and `go.mod` reflects that stable toolchain.
- The port will be done in the same repository so docs, fixtures, and golden
  outputs cannot drift across repos.
- The Python CLI remains authoritative until the final cutover task.
- The first Go pass should be stdlib-first: `encoding/json`, `flag` or a thin
  local subcommand dispatcher, `filepath.WalkDir`, `regexp`, `sort`, `testing`,
  and `os/exec` where needed for parity tests.
- Third-party Go dependencies require a specific reason recorded in the task that
  adds them. Candidate reasons include materially better CLI help parity or a
  parser dependency needed to preserve existing behavior.
- All generated JSON comparisons must normalize machine-local absolute paths
  before asserting equality.
- Human output must be smoke-tested for key strings, but JSON output is the
  compatibility surface that needs strict parity.
- The Python source scanner currently uses `ast` and recognizes only documented
  literal patterns. The Go port should match the documented behavior first and
  add a parser dependency only if fixture or golden parity proves regex scanning
  is insufficient.
- Plan and spec lifecycle frontmatter must stay accurate. This spec starts as
  `draft`, moves to `in-progress` when implementation starts, and moves to
  `complete` only after the cutover decision is made.

## Task Breakdown

### Task 1: Prepare Go Toolchain And Module Skeleton

**Objective**

Install or expose the Go toolchain, add the initial Go module, and create a
side-by-side executable that does not change the existing Python launcher.

**Files**

- Create: `go.mod`
- Create: `cmd/envdiff/main.go`
- Create: `internal/cli/cli.go`
- Create: `internal/version/version.go`
- Create: `internal/testutil/paths.go`
- Modify: `.gitignore` if Go build artifacts need to be ignored.

**Dependencies**

None

**Implementation Steps**

1. Install Go or update PATH so `go version` works locally.
2. Record the installed Go version in the implementation notes for the task.
3. Initialize `go.mod` with a module path for this repo and a `go` directive that
   matches the installed stable toolchain.
4. Add `cmd/envdiff/main.go` that calls the local CLI package.
5. Add a minimal CLI dispatcher that supports `--help` and exits non-zero for
   unimplemented subcommands.
6. Add `internal/testutil/paths.go` helpers for locating repo fixtures from Go
   tests without depending on the caller's current directory.
7. Keep the root `./envdiff` launcher unchanged.

**Verification**

- Run: `go version`
- Expect: prints the local Go version.
- Run: `go test ./...`
- Expect: all current Go packages compile and pass.
- Run: `go run ./cmd/envdiff --help`
- Expect: exits `0` and lists the planned command names.
- Run: `./envdiff --help`
- Expect: still runs the existing Python CLI.

**Done When**

- The repo has a compiling Go module.
- `go run ./cmd/envdiff --help` works.
- The Python launcher and Python test suite are untouched.

### Task 2: Add Python-Oracle Golden Test Harness

**Objective**

Capture stable Python CLI outputs for representative command paths and create a
repeatable comparison harness for the Go implementation.

**Files**

- Create: `scripts/update_go_golden.py`
- Create: `tests/golden/README.md`
- Create: `tests/golden/json/compare-basic.json`
- Create: `tests/golden/json/matrix-basic.json`
- Create: `tests/golden/json/matrix-show-all.json`
- Create: `tests/golden/json/scan-simple-repo.json`
- Create: `tests/golden/json/scan-workflow-repo.json`
- Create: `tests/golden/json/generate-simple-repo.json`
- Create: `tests/golden/json/generate-simple-repo-annotated.json`
- Create: `tests/golden/json/doctor-project.json`
- Create: `tests/golden/json/doctor-project-baseline.json`
- Create: `internal/testutil/golden.go`

**Dependencies**

- Task 1

**Implementation Steps**

1. Add a Python script that invokes `./envdiff` for the selected JSON command
   cases.
2. Normalize absolute paths in generated JSON to stable placeholders before
   writing golden files.
3. Include at least one golden for suppression metadata by generating a baseline
   in a controlled temporary directory, then normalizing the temporary path.
4. Document when to regenerate goldens and how path normalization works.
5. Add Go test helpers to load golden JSON and compare normalized JSON values,
   not raw whitespace.
6. Do not use goldens for human output beyond smoke strings.

**Verification**

- Run: `uv run python scripts/update_go_golden.py --check`
- Expect: exits `0` after confirming committed goldens match Python output.
- Run: `uv run pytest -q`
- Expect: existing Python tests still pass.
- Run: `go test ./...`
- Expect: Go helper packages pass, even if no command parity tests are active yet.

**Done When**

- Golden JSON files exist for all command families.
- Golden generation is repeatable on a different local checkout path.
- The Go test harness can compare normalized JSON structures.

### Task 3: Port Models, JSON Rendering, And Ordering

**Objective**

Create Go data structures and serialization helpers that match the Python
`models.py`, `render/json.py`, and `utils/ordering.py` contracts.

**Files**

- Create: `internal/model/model.go`
- Create: `internal/render/json.go`
- Create: `internal/order/order.go`
- Create: `internal/model/model_test.go`
- Create: `internal/render/json_test.go`
- Create: `internal/order/order_test.go`
- Reference: `src/envdiff/models.py`
- Reference: `src/envdiff/render/json.py`
- Reference: `src/envdiff/utils/ordering.py`
- Reference: `docs/system/JSON_SCHEMA.md`

**Dependencies**

- Task 1
- Task 2

**Implementation Steps**

1. Define Go structs for all public data shapes:
   `CommandMeta`, `Location`, `EnvVarDefinition`, `EnvVarUsage`,
   `EnvVarContract`, `ResolutionDecision`, `RepoScanResult`, `BaselineEntry`,
   `BaselineSnapshot`, `Finding`, `SummaryCounts`, and `JsonEnvelope`.
2. Add JSON struct tags that preserve exact field names.
3. Preserve null vs empty collection behavior where schema version `1` expects it.
4. Implement schema version `1` as a single constant.
5. Implement deterministic sort helpers for definitions, usages, contracts,
   findings, and strings.
6. Implement JSON rendering with stable indentation and map-key ordering where
   command `data` uses map-like payloads.
7. Add focused tests against small inline examples and selected golden files.

**Verification**

- Run: `go test ./internal/model ./internal/render ./internal/order`
- Expect: model, rendering, and ordering tests pass.
- Run: `go test ./...`
- Expect: all Go packages pass.
- Run: `uv run pytest tests/test_models.py -q`
- Expect: Python model contract tests still pass.

**Done When**

- Go model output can represent every documented JSON shape.
- Ordering helpers are available to all later parser and analyzer tasks.
- Schema version `1` is represented in one Go constant.

### Task 4: Port Dotenv Parsing And Value Normalization

**Objective**

Implement dotenv parsing and normalized value classification in Go with fixture
coverage matching the Python parser.

**Files**

- Create: `internal/dotenv/parse.go`
- Create: `internal/dotenv/parse_test.go`
- Create: `internal/normalize/value.go`
- Create: `internal/normalize/value_test.go`
- Reference: `src/envdiff/parsers/dotenv.py`
- Reference: `src/envdiff/utils/normalize.py`
- Reuse: `tests/fixtures/dotenv/basic.env`
- Reuse: `tests/fixtures/dotenv/duplicates.env`
- Reuse: `tests/fixtures/dotenv/malformed.env`

**Dependencies**

- Task 3

**Implementation Steps**

1. Port dotenv line parsing for comments, blank lines, key/value definitions,
   single-quoted values, double-quoted values, duplicates, and malformed syntax
   warnings.
2. Preserve line numbers, file paths, duplicate flags, parse warnings, and
   normalized value kinds.
3. Preserve the existing conservative syntax boundary rather than adding broad
   dotenv dialect support.
4. Add table-driven tests for existing fixtures and inline edge cases.
5. Compare parser output to Python output where practical through normalized JSON.

**Verification**

- Run: `go test ./internal/dotenv ./internal/normalize`
- Expect: dotenv and normalization tests pass.
- Run: `uv run pytest tests/test_dotenv.py tests/test_utils.py -q`
- Expect: Python parser and normalization tests still pass.

**Done When**

- Go parses all committed dotenv fixtures with matching definitions and warnings.
- Duplicate handling and value-kind normalization match Python behavior.
- Unsupported dotenv syntax produces warnings rather than crashes.

### Task 5: Port Source Scanners

**Objective**

Implement Go scanners for Python env usage, Docker Compose interpolation, and
GitHub Actions expressions.

**Files**

- Create: `internal/parsers/python.go`
- Create: `internal/parsers/python_test.go`
- Create: `internal/parsers/compose.go`
- Create: `internal/parsers/compose_test.go`
- Create: `internal/parsers/github_actions.go`
- Create: `internal/parsers/github_actions_test.go`
- Reference: `src/envdiff/parsers/python_ast.py`
- Reference: `src/envdiff/parsers/compose.py`
- Reference: `src/envdiff/parsers/github_actions.py`
- Reuse: `tests/fixtures/python/sample_app.py`
- Reuse: `tests/fixtures/python/unsupported.py`
- Reuse: `tests/fixtures/compose/docker-compose.yml`
- Reuse: `tests/fixtures/github_actions/deploy.yml`

**Dependencies**

- Task 3

**Implementation Steps**

1. Implement Python scanning for documented literal patterns:
   `os.environ["X"]`, `os.getenv("X")`, and `os.getenv("X", "default")`.
2. Preserve requiredness behavior:
   `os.environ[...]` is `required`, `os.getenv("X")` is `optional`, and
   `os.getenv("X", "default")` is `optional_with_default`.
3. Record file path, line number, usage kind, requiredness, default value, and
   source type.
4. Implement Compose interpolation scanning for `${VAR}` and `${VAR:-default}`.
5. Implement GitHub Actions expression scanning for `secrets.NAME` and
   `vars.NAME`, including `||` default extraction.
6. Add tests for documented positive patterns and unsupported patterns.
7. If line-oriented scanning cannot preserve documented behavior, add a focused
   parser dependency in this task and record the reason in `go.mod`.

**Verification**

- Run: `go test ./internal/parsers`
- Expect: all scanner tests pass.
- Run: `uv run pytest tests/test_python_ast.py tests/test_compose.py tests/test_github_actions.py -q`
- Expect: Python scanner tests still pass.
- Run: `go test ./...`
- Expect: all Go packages pass.

**Done When**

- All existing source/config fixtures produce matching usage records.
- Unsupported patterns fail soft.
- No parser dependency is added unless parity requires it.

### Task 6: Port Repo Traversal, Resolution, And Contract Aggregation

**Objective**

Implement repo scanning in Go so definitions, usages, resolutions, and contracts
match the Python scanner.

**Files**

- Create: `internal/paths/walk.go`
- Create: `internal/paths/walk_test.go`
- Create: `internal/analyzers/scan.go`
- Create: `internal/analyzers/scan_test.go`
- Reference: `src/envdiff/utils/paths.py`
- Reference: `src/envdiff/analyzers/scan.py`
- Reuse: `tests/fixtures/repos/simple_repo`
- Reuse: `tests/fixtures/repos/workflow_repo`
- Reuse: `tests/fixtures/doctor/project`

**Dependencies**

- Task 4
- Task 5

**Implementation Steps**

1. Port deterministic repository traversal and ignore behavior.
2. Port nearest `.env` and `.env.example` resolution for usage files.
3. Dispatch files to dotenv, Python, Compose, and GitHub Actions parsers.
4. Aggregate definitions and usages by variable name.
5. Infer contract requiredness and statuses using Python-equivalent rules.
6. Preserve resolution notes, warning sorting, and contract ordering.
7. Add golden comparison tests for `scan` on simple and workflow fixture repos.

**Verification**

- Run: `go test ./internal/paths ./internal/analyzers -run 'Walk|Scan'`
- Expect: traversal and scan tests pass.
- Run: `go run ./cmd/envdiff scan tests/fixtures/repos/simple_repo --json`
- Expect: exits `0` and normalized JSON matches `tests/golden/json/scan-simple-repo.json`.
- Run: `uv run pytest tests/test_scan.py -q`
- Expect: Python scan tests still pass.

**Done When**

- Go scan output matches Python goldens for committed fixtures.
- Repo traversal is deterministic.
- Resolution behavior is covered before doctor/generate rely on it.

### Task 7: Port Compare And Matrix Commands

**Objective**

Ship Go implementations for file-to-file and multi-file dotenv comparisons.

**Files**

- Create: `internal/analyzers/compare.go`
- Create: `internal/analyzers/compare_test.go`
- Create: `internal/analyzers/matrix.go`
- Create: `internal/analyzers/matrix_test.go`
- Modify: `internal/cli/cli.go`
- Reference: `src/envdiff/analyzers/compare.py`
- Reference: `src/envdiff/analyzers/matrix.py`
- Reuse: `tests/fixtures/compare/left.env`
- Reuse: `tests/fixtures/compare/right.env`
- Reuse: `tests/fixtures/matrix/a.env`
- Reuse: `tests/fixtures/matrix/b.env`
- Reuse: `tests/fixtures/matrix/c.env`

**Dependencies**

- Task 4
- Task 3

**Implementation Steps**

1. Port compare result construction for missing keys, duplicate keys, differing
   value kinds, and warnings.
2. Port matrix result construction for presence, missing files, duplicates,
   value-kind maps, inconsistent counts, and `--show-all`.
3. Wire `compare` and `matrix` into the Go CLI for JSON and human modes.
4. Implement argument validation for `matrix` requiring at least two paths.
5. Add command-level golden JSON tests.
6. Add human-output smoke tests for key headings, not byte-for-byte formatting.

**Verification**

- Run: `go test ./internal/analyzers -run 'Compare|Matrix'`
- Expect: compare and matrix analyzer tests pass.
- Run: `go run ./cmd/envdiff compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env --json`
- Expect: normalized JSON matches `tests/golden/json/compare-basic.json`.
- Run: `go run ./cmd/envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env --json`
- Expect: normalized JSON matches `tests/golden/json/matrix-basic.json`.
- Run: `go run ./cmd/envdiff matrix tests/fixtures/matrix/a.env`
- Expect: exits non-zero and reports that matrix requires at least two dotenv files.

**Done When**

- `compare` and `matrix` work through the Go CLI.
- JSON output matches Python goldens.
- Argument validation and human smoke behavior are covered.

### Task 8: Port Generate Command

**Objective**

Implement `.env.example` generation, annotation, output writing, and drift check
behavior in Go.

**Files**

- Create: `internal/analyzers/generate.go`
- Create: `internal/analyzers/generate_test.go`
- Modify: `internal/cli/cli.go`
- Reference: `src/envdiff/analyzers/generate.py`
- Reuse: `tests/fixtures/repos/simple_repo`

**Dependencies**

- Task 6
- Task 3

**Implementation Steps**

1. Port variable selection and ordering from scan contracts.
2. Port plain generated dotenv output.
3. Port annotated output comments and default notes.
4. Port `--output` file writing behavior.
5. Port `--check` target resolution and drift detection.
6. Wire `generate` into the Go CLI for JSON and human modes.
7. Preserve exit code `2` when `--check` detects drift.
8. Add golden JSON tests for plain and annotated modes.

**Verification**

- Run: `go test ./internal/analyzers -run Generate`
- Expect: generate analyzer tests pass.
- Run: `go run ./cmd/envdiff generate tests/fixtures/repos/simple_repo`
- Expect: prints `DATABASE_URL=`, `REDIS_URL=`, and `DEBUG=` in Python-equivalent order.
- Run: `go run ./cmd/envdiff generate tests/fixtures/repos/simple_repo --json`
- Expect: normalized JSON matches `tests/golden/json/generate-simple-repo.json`.
- Run: `go run ./cmd/envdiff generate tests/fixtures/repos/simple_repo --check`
- Expect: exits `2` when the fixture template has drift, matching Python behavior.
- Run: `uv run pytest tests/test_generate.py -q`
- Expect: Python generate tests still pass.

**Done When**

- Go `generate` matches Python behavior for stdout, output file, JSON, and check
  modes.
- Drift check exit behavior is tested.
- No automatic repo rewrite behavior is introduced.

### Task 9: Port Doctor Findings, Suppressions, And Baselines

**Objective**

Implement `doctor` validation in Go, including all deterministic and heuristic
findings plus baseline and ignore-file suppression.

**Files**

- Create: `internal/analyzers/doctor.go`
- Create: `internal/analyzers/doctor_test.go`
- Create: `internal/analyzers/aliases.go`
- Create: `internal/analyzers/aliases_test.go`
- Create: `internal/analyzers/secrets.go`
- Create: `internal/analyzers/secrets_test.go`
- Create: `internal/analyzers/baseline.go`
- Create: `internal/analyzers/baseline_test.go`
- Modify: `internal/cli/cli.go`
- Reference: `src/envdiff/analyzers/doctor.py`
- Reference: `src/envdiff/analyzers/aliases.py`
- Reference: `src/envdiff/analyzers/secrets.py`
- Reference: `src/envdiff/analyzers/baseline.py`
- Reference: `docs/system/FINDING_CODES.md`
- Reuse: `tests/fixtures/doctor/project`

**Dependencies**

- Task 6
- Task 3

**Implementation Steps**

1. Port missing required, missing optional, unused, undocumented, template skew,
   and duplicate findings.
2. Port alias candidate heuristics with the same conservative reasons and
   ordering.
3. Port secret-like and placeholder-like value checks.
4. Port suppression key generation exactly.
5. Port baseline snapshot loading and writing.
6. Port explicit `--ignore-file` and default `.envdiffignore` behavior.
7. Wire `doctor` into the Go CLI for JSON and human modes.
8. Preserve `--fail-on error|warning|info` threshold behavior and invalid
   threshold handling.
9. Add golden JSON tests for active findings and suppressed findings.

**Verification**

- Run: `go test ./internal/analyzers -run 'Doctor|Alias|Secret|Baseline'`
- Expect: doctor, heuristic, and baseline tests pass.
- Run: `go run ./cmd/envdiff doctor tests/fixtures/doctor/project --json`
- Expect: normalized JSON matches `tests/golden/json/doctor-project.json`.
- Run: `go run ./cmd/envdiff doctor tests/fixtures/doctor/project --fail-on warning`
- Expect: exits `2` and includes `ENV001`.
- Run: `go run ./cmd/envdiff doctor tests/fixtures/doctor/project --fail-on debug`
- Expect: exits non-zero and reports allowed severities.
- Run: `uv run pytest tests/test_doctor.py tests/test_aliases.py tests/test_secrets.py tests/test_baseline.py -q`
- Expect: Python doctor-related tests still pass.

**Done When**

- Go `doctor` emits all current finding codes with matching severities and
  suppression keys.
- Baseline and ignore-file suppression metadata matches Python JSON.
- Exit-code behavior matches Python for thresholds and write-baseline mode.

### Task 10: Port Human Rendering And CLI UX

**Objective**

Provide human-readable Go output that preserves the current user-facing command
signals without treating cosmetic whitespace as a machine contract.

**Files**

- Create: `internal/render/human.go`
- Create: `internal/render/human_test.go`
- Modify: `internal/cli/cli.go`
- Reference: `src/envdiff/render/human.py`
- Reference: `tests/test_render_human.py`
- Reference: `tests/test_cli_smoke.py`

**Dependencies**

- Task 7
- Task 8
- Task 9

**Implementation Steps**

1. Port human summaries for compare, scan, matrix, generate, and doctor.
2. Preserve key strings used by smoke tests, including headings and summary labels.
3. Implement help output with command names and option descriptions.
4. Preserve no-args help behavior.
5. Add Go CLI smoke tests for every command in human and JSON modes.
6. Avoid introducing a third-party CLI framework unless stdlib dispatch prevents
   acceptable help output or validation behavior.

**Verification**

- Run: `go test ./internal/render ./internal/cli`
- Expect: human rendering and CLI smoke tests pass.
- Run: `go run ./cmd/envdiff --help`
- Expect: exits `0` and lists `compare`, `generate`, `matrix`, `scan`, and `doctor`.
- Run: `go run ./cmd/envdiff scan tests/fixtures/repos/simple_repo`
- Expect: exits `0` and includes `Contracts: 3`.
- Run: `go run ./cmd/envdiff doctor tests/fixtures/doctor/project --fail-on warning`
- Expect: exits `2` and includes summary counts.

**Done When**

- Human output is usable in a terminal.
- All command names and common options are discoverable through help.
- CLI smoke tests cover both JSON and human paths.

### Task 11: Add Cross-Implementation Parity Gate

**Objective**

Create a single local command path that proves Go matches Python for the selected
contract-critical cases before any launcher switch.

**Files**

- Create: `scripts/check_go_parity.py`
- Create: `tests/test_go_parity.py`
- Modify: `README.md`
- Modify: `docs/system/OPERATIONS.md`
- Modify: `docs/system/JSON_SCHEMA.md` only if the verification procedure needs a
  non-contract-changing note.

**Dependencies**

- Task 7
- Task 8
- Task 9
- Task 10

**Implementation Steps**

1. Add a Python parity script that runs Python `./envdiff` and Go
   `go run ./cmd/envdiff` for the golden command matrix.
2. Normalize absolute paths, temporary paths, and platform-specific path
   separators before comparison.
3. Compare JSON structurally, not as raw strings.
4. Check key human-output smoke strings for representative commands.
5. Check exit codes for success, matrix single-argument failure, doctor threshold
   failure, invalid fail-on value, and generate drift failure.
6. Add a pytest wrapper that skips clearly when `go` is unavailable, so Python-only
   CI remains usable until CI is updated.
7. Document the parity command in operations docs.

**Verification**

- Run: `uv run python scripts/check_go_parity.py`
- Expect: exits `0` after all JSON, human smoke, and exit-code cases pass.
- Run: `uv run python -m pytest -q tests/test_go_parity.py`
- Expect: passes when Go is installed or skips with an explicit toolchain message.
- Run: `uv run pytest -q`
- Expect: existing Python suite remains green.
- Run: `go test ./...`
- Expect: all Go tests pass.

**Done When**

- There is one documented parity command.
- JSON parity covers every command family.
- Exit-code parity covers the current public behavior.

### Task 12: Update CI For Dual Python And Go Validation

**Objective**

Teach CI to run Go tests and the parity gate without weakening the existing
Python gate.

**Files**

- Modify: `.github/workflows/ci.yml`
- Modify: `AGENTS.md`
- Modify: `README.md`
- Modify: `docs/system/OPERATIONS.md`

**Dependencies**

- Task 11

**Implementation Steps**

1. Add a pinned Go setup step to CI following the repo's existing pinned-action
   style.
2. Add `go test ./...` to CI.
3. Add the parity script to CI after both Python dependencies and Go are available.
4. Keep existing `ruff`, format, dead-code, and pytest jobs.
5. Update command quickstarts to include Go verification commands.
6. State that `go` is now a dev prerequisite only after CI uses it.

**Verification**

- Run: `uv run ruff check .`
- Expect: exits `0`.
- Run: `uv run ruff format --check .`
- Expect: exits `0`.
- Run: `uv run pytest -q`
- Expect: exits `0`.
- Run: `go test ./...`
- Expect: exits `0`.
- Run: `uv run python scripts/check_go_parity.py`
- Expect: exits `0`.

**Done When**

- CI exercises both Python and Go.
- Local docs list the same verification commands that CI runs.
- Go becomes a documented dev prerequisite.

### Task 13: Switch The Local Launcher After Parity

**Objective**

Make the Go implementation the default `./envdiff` path only after all parity
gates pass.

**Files**

- Modify: `envdiff`
- Create: `scripts/envdiff-python` or `scripts/envdiff-go` if a temporary fallback
  launcher is useful.
- Modify: `README.md`
- Modify: `docs/system/OPERATIONS.md`
- Modify: `docs/project/ROADMAP.md`
- Modify: `AGENTS.md`

**Dependencies**

- Task 12

**Implementation Steps**

1. Re-run the full Python, Go, and parity gates before editing the launcher.
2. Decide whether `./envdiff` should execute a checked-in built binary, run
   `go run`, or dispatch to a locally built binary path. Prefer a simple dev
   launcher for this repo unless packaging requirements are added.
3. Preserve an explicit way to run the Python implementation during the transition.
4. Update docs to describe the Go default path and Python legacy/oracle path.
5. Update roadmap status to record the Go port cutover.
6. Re-run all verification commands after the launcher switch.

**Verification**

- Run: `./envdiff --help`
- Expect: exits `0` through the Go implementation and lists all commands.
- Run: `./envdiff scan tests/fixtures/repos/simple_repo --json`
- Expect: normalized JSON matches the Python golden for simple repo scan.
- Run: `uv run pytest -q`
- Expect: Python suite still passes or intentionally updated launcher tests pass.
- Run: `go test ./...`
- Expect: exits `0`.
- Run: `uv run python scripts/check_go_parity.py`
- Expect: exits `0`.

**Done When**

- `./envdiff` defaults to Go.
- A documented Python fallback remains available during the transition.
- Docs and roadmap match the new default implementation.

### Task 14: Decide Python Legacy Removal Or Retention

**Objective**

Make an explicit post-cutover decision about whether the Python implementation
stays as a test oracle, moves to a legacy folder, or is removed.

**Files**

- Modify: `src/envdiff/**` if moving or removing Python code.
- Modify: `pyproject.toml` if Python packaging changes.
- Modify: `uv.lock` if Python dependencies change.
- Modify: `tests/**` if Python tests become legacy-only or are removed.
- Modify: `README.md`
- Modify: `docs/system/OPERATIONS.md`
- Modify: `docs/project/ROADMAP.md`
- Modify: `AGENTS.md`

**Dependencies**

- Task 13

**Implementation Steps**

1. Review how useful the Python implementation still is as an oracle after the
   launcher switch.
2. Choose one of three explicit outcomes:
   - Retain Python as a legacy oracle for one release window.
   - Move Python under a clearly named legacy path.
   - Remove Python package code and keep only generated goldens.
3. Update tests and CI to match the chosen outcome.
4. Update docs so setup commands do not imply two active implementations unless
   both are intentionally supported.
5. Run the full remaining gate after cleanup.

**Verification**

- Run: `./envdiff --help`
- Expect: Go CLI works.
- Run: `go test ./...`
- Expect: exits `0`.
- Run: `uv run pytest -q` if Python tests remain.
- Expect: exits `0`, or the command is no longer part of the documented gate.
- Run: `uv run ruff check .` if Python code remains.
- Expect: exits `0`, or the command is no longer part of the documented gate.

**Done When**

- There is one documented ownership model for Python after cutover.
- CI and docs match that ownership model.
- No stale references imply unsupported Python behavior.

## Risks And Mitigations

- Risk: JSON field ordering, nullability, or nested shape drifts during the Go
  rewrite.
  Mitigation: add Python-generated golden JSON before analyzer ports and compare
  normalized structures for every command family.

- Risk: Absolute paths make committed goldens machine-specific.
  Mitigation: normalize repo roots, temp directories, and path separators in both
  golden generation and parity checks.

- Risk: The Python scanner's AST behavior is accidentally broadened or narrowed
  by a simpler Go implementation.
  Mitigation: constrain parity to documented literal syntax first, add fixtures
  for edge cases, and add a Go parser dependency only if documented behavior
  cannot be preserved with a small scanner.

- Risk: The port changes product behavior while trying to improve internals.
  Mitigation: keep this pass parity-only and record any desired product changes
  as future issues or roadmap entries.

- Risk: Maintaining Python and Go side by side creates duplicate implementation
  burden.
  Mitigation: keep side-by-side mode temporary, with an explicit cutover task and
  a final legacy-removal-or-retention decision.

- Risk: CI becomes slow or brittle after adding parity checks.
  Mitigation: keep fixture scope small, avoid network access, and split pure Go
  unit tests from cross-implementation parity tests.

- Risk: The local environment lacks Go at the start of implementation.
  Mitigation: make toolchain installation Task 1 and skip parity pytest cleanly
  until Go is installed.

- Risk: Human output parity becomes a distraction.
  Mitigation: make JSON strict and human output smoke-based, preserving key user
  signals without byte-for-byte Rich formatting.

## Verification Matrix

| Requirement | Proof command | Expected signal |
| --- | --- | --- |
| Existing Python behavior remains intact during the side-by-side phase | `uv run pytest -q` | All Python tests pass |
| Python lint and format gate remains intact | `uv run ruff check . && uv run ruff format --check .` | Both commands exit `0` |
| Go module compiles and unit tests pass | `go test ./...` | All Go packages pass |
| Golden JSON can be regenerated reproducibly | `uv run python scripts/update_go_golden.py --check` | Command exits `0` with no golden drift |
| Go compare JSON matches Python | `go run ./cmd/envdiff compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env --json` | Normalized JSON matches `tests/golden/json/compare-basic.json` |
| Go matrix JSON matches Python | `go run ./cmd/envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env --json` | Normalized JSON matches `tests/golden/json/matrix-basic.json` |
| Go scan JSON matches Python | `go run ./cmd/envdiff scan tests/fixtures/repos/simple_repo --json` | Normalized JSON matches `tests/golden/json/scan-simple-repo.json` |
| Go generate JSON matches Python | `go run ./cmd/envdiff generate tests/fixtures/repos/simple_repo --json` | Normalized JSON matches `tests/golden/json/generate-simple-repo.json` |
| Go doctor JSON matches Python | `go run ./cmd/envdiff doctor tests/fixtures/doctor/project --json` | Normalized JSON matches `tests/golden/json/doctor-project.json` |
| Exit code behavior matches Python | `uv run python scripts/check_go_parity.py` | Success, threshold failure, invalid argument, and drift cases match |
| Core remains offline and deterministic | `uv run python scripts/check_go_parity.py && go test ./...` | Repeated runs produce the same normalized outputs |
| Launcher switch is safe | `./envdiff --help && ./envdiff scan tests/fixtures/repos/simple_repo --json` | Go default launcher works and scan JSON matches golden output |

## Handoff

1. Execute in this session, task by task.
2. Open a separate execution session.
3. Refine this spec before implementation.
