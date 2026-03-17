# Operations

Operational notes for local development and verification.

## Setup

Install dependencies and dev tools:

```bash
uv sync --extra dev
```

## Key Commands

- Run CLI help: `uv run python -m envdiff.cli --help`
- Run tests: `uv run pytest -q`
- Lint: `uv run ruff check .`

## Runnable Fixture Projects

Use these paths as local demo environments:

- `tests/fixtures/repos/simple_repo`
- `tests/fixtures/repos/monorepo`
- `tests/fixtures/doctor/project`

Examples:

```bash
uv run python -m envdiff.cli scan tests/fixtures/repos/simple_repo --json
uv run python -m envdiff.cli scan tests/fixtures/repos/monorepo --json
uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --fail-on warning
uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --write-baseline .envdiff.baseline.json
uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --baseline .envdiff.baseline.json
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

- compare / scan / doctor
- repo-local `.env` and `.env.example` resolution
- deterministic JSON output
- alias, secret-like, and placeholder-like heuristics
- baseline snapshots and ignore-based suppression
- grouped doctor summaries in human output

What is still open:

- finding-noise reduction beyond the initial heuristic pass
- JSON schema documentation and finding-code reference docs
- matrix and generate workflows
- broader parser coverage
