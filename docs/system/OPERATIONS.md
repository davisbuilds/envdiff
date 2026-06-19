# Operations

Operational notes for local development and verification.

## Setup

envdiff is a Go program; Go `1.26.4` is the only requirement. The lint stack
uses `golangci-lint` (v2) and `gofumpt`.

## Key Commands

- Run CLI help: `./envdiff --help`
- Run tests: `go test ./...`
- Lint: `golangci-lint run ./...`
- Regenerate JSON goldens from Go: `ENVDIFF_UPDATE_GOLDENS=1 go test ./...`

`./envdiff` is the launcher: it builds and caches `bin/envdiff`, rebuilding only
when sources change. `go build -o bin/envdiff ./cmd/envdiff` builds it directly.

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
- `2`: findings met or exceeded `--fail-on` (case-insensitive), `generate --check`
  drift (including with `--json`), or a usage/validation error such as
  single-file `matrix`

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
- Go implementation under `cmd/envdiff/` and `internal/`
- `./envdiff` launcher that builds and caches `bin/envdiff`
- JSON goldens generated from Go (`ENVDIFF_UPDATE_GOLDENS=1 go test ./...`)

What is still open:

- finding-noise reduction beyond the initial heuristic pass
- broader parser coverage
