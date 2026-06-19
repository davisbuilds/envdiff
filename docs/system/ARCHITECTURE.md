# Architecture

## High-Level Flow

`envdiff` has five top-level workflows:

1. `compare`: parse two dotenv files and compute missing keys, duplicates, and value-kind differences.
2. `generate`: infer a repo-local `.env.example` candidate and optionally check drift.
3. `matrix`: compare multiple dotenv files across presence, value-kind, and duplicate signals.
4. `scan`: walk a repo, parse supported definition files, scan supported source files, and aggregate contracts.
5. `doctor`: run deterministic validation on the scan result and emit structured findings.

The current implementation is intentionally local-first and deterministic. There are no network dependencies in core analysis.

## Implementation Ownership

envdiff is a Go program: `cmd/envdiff/` plus the packages under `internal/`,
run locally through the `./envdiff` launcher. It began as a port of a Python
implementation that served as a transitional parity oracle; that oracle has been
retired and Go is the sole source of truth (see
`docs/plans/2026-06-18-go-source-of-truth-migration.md`).

## CLI Layer

`cmd/envdiff/main.go` delegates to `internal/cli`, which exposes the current
command surface:

```text
compare, generate, matrix, scan, doctor
```

Each command supports a human-oriented terminal rendering path and a stable JSON path.

## Analyzer Layer

`internal/analyzers/` contains the active Go product logic:

- `compare.go`: dotenv file-to-file comparison
- `generate.go`: inferred `.env.example` generation and drift checks
- `matrix.go`: multi-file dotenv comparison
- `scan.go`: repository traversal, parser dispatch, contract aggregation, and repo-local resolution; files are parsed in a bounded worker pool and merged in file order, so output stays deterministic
- `doctor.go`: contract validation and finding generation
- `aliases.go`: low-confidence, explainable naming-drift heuristics
- `secrets.go`: conservative secret-like and placeholder-like checks for committed `.env` values

## Parser Layer

`internal/dotenv/`, `internal/parsers/`, and `internal/paths/` contain the
currently supported input surface:

- `internal/dotenv/parse.go`: `.env` and `.env.example` parsing with duplicate preservation and warnings
- `internal/parsers/python.go`: `os.environ[...]` and `os.getenv(...)`
- `internal/parsers/compose.go`: Docker Compose `${VAR}` interpolation
- `internal/parsers/github_actions.go`: workflow expression scanning for `secrets.*` and `vars.*`

## Model Layer

`internal/model/` defines the active shared data contract:

- `EnvVarDefinition`
- `EnvVarUsage`
- `EnvVarContract`
- `Finding`
- `RepoScanResult`
- `JsonEnvelope`

The JSON envelope is versioned and intended to remain a stable machine contract.
JSON is encoded without HTML-escaping, so `<`, `>`, and `&` appear literally in
values rather than as `\u00XX`.

## Utilities

`internal/order/` and `internal/normalize/` provide deterministic helpers for:

- value normalization
- stable ordering
- repo traversal
- nearest file resolution

## Directory Map

```text
envdiff                         # Local Go launcher wrapper
cmd/envdiff/                    # Go CLI entrypoint
internal/analyzers/             # Comparison, scan, doctor, alias, and secret logic
internal/dotenv/                # Dotenv parser
internal/parsers/               # Python, Compose, and GitHub Actions scanners
internal/lines/                 # Cap-free file-to-lines reader (ScanLines semantics)
internal/render/                # Human and JSON renderers
internal/model/                 # JSON schema and domain models
internal/order/                 # Deterministic ordering helpers
internal/normalize/             # Value normalization helpers
internal/paths/                 # Repo traversal and nearest-file helpers
internal/version/               # Schema and build version constants
tests/fixtures/                 # Runnable example repos and file fixtures
tests/golden/json/              # Rendered-output goldens (generated from Go)
docs/system/                    # Architecture, features, operations
                                # JSON contract and finding-code references
docs/project/                   # Spec, roadmap, and backlog
docs/archive/                   # Archived plans and research
```
