# Operations

Operational notes for local development and verification.

## Setup

Install dependencies and dev tools:

```bash
uv sync --extra dev
```

Go `1.26.4` is also required for the side-by-side Go implementation, Go tests,
and Python/Go parity gate.

## Key Commands

- Run CLI help: `./envdiff --help`
- Run Python fallback help: `scripts/envdiff-python --help`
- Run tests: `uv run pytest -q`
- Run Go tests: `go test ./...`
- Lint: `uv run ruff check .`
- Run Go parity gate: `uv run python scripts/check_go_parity.py`

During the Go cutover, `./envdiff` is the Go launcher. The Python CLI remains
available as `scripts/envdiff-python` and is still used as the parity oracle.

## Runnable Fixture Projects

Use these paths as local demo environments:

- `tests/fixtures/repos/workflow_repo`
- `tests/fixtures/matrix`
- `tests/fixtures/repos/simple_repo`
- `tests/fixtures/repos/monorepo`
- `tests/fixtures/doctor/project`

Examples:

```bash
./envdiff generate tests/fixtures/repos/simple_repo --annotate
./envdiff generate tests/fixtures/repos/simple_repo --check
./envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env
./envdiff scan tests/fixtures/repos/workflow_repo --json
./envdiff scan tests/fixtures/repos/simple_repo --json
./envdiff scan tests/fixtures/repos/monorepo --json
./envdiff doctor tests/fixtures/doctor/project --fail-on warning
./envdiff doctor tests/fixtures/doctor/project --write-baseline .envdiff.baseline.json
./envdiff doctor tests/fixtures/doctor/project --baseline .envdiff.baseline.json
```

## Exit Codes

- `0`: command completed without findings at or above threshold
- `1`: execution or parsing failure
- `2`: findings met or exceeded `--fail-on`

## Current Constraints

- repo-local analysis only
- no user-shell startup files
- no env loading or process injection
- no external secret APIs

## Implementation Status

What is implemented:

- compare / generate / matrix / scan / doctor
- repo-local `.env` and `.env.example` resolution
- deterministic JSON output
- alias, secret-like, and placeholder-like heuristics
- baseline snapshots and ignore-based suppression
- grouped doctor summaries in human output
- side-by-side Go implementation under `cmd/envdiff/` and `internal/`
- Go default launcher at `./envdiff`
- Python fallback launcher at `scripts/envdiff-python`
- Python/Go parity script for contract-critical fixture cases

What is still open:

- final Python legacy retention/removal decision
- finding-noise reduction beyond the initial heuristic pass
- broader parser coverage
