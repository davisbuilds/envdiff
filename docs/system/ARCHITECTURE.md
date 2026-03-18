# Architecture

## High-Level Flow

`envdiff` currently has five top-level workflows:

1. `compare`: parse two dotenv files and compute missing keys, duplicates, and value-kind differences.
2. `generate`: infer a repo-local `.env.example` candidate and optionally check drift.
3. `matrix`: compare multiple dotenv files across presence, value-kind, and duplicate signals.
4. `scan`: walk a repo, parse supported definition files, scan supported source files, and aggregate contracts.
5. `doctor`: run deterministic validation on the scan result and emit structured findings.

The current implementation is intentionally local-first and deterministic. There are no network dependencies in core analysis.

## CLI Layer

`src/cli.py` exposes the current command surface:

```text
compare, generate, matrix, scan, doctor
```

Each command supports a human-oriented terminal rendering path and a stable JSON path.

## Analyzer Layer

`src/analyzers/` contains the main product logic:

- `compare.py`: dotenv file-to-file comparison
- `generate.py`: inferred `.env.example` generation and drift checks
- `matrix.py`: multi-file dotenv comparison
- `scan.py`: repository traversal, parser dispatch, contract aggregation, and repo-local resolution
- `doctor.py`: contract validation and finding generation
- `aliases.py`: low-confidence, explainable naming-drift heuristics
- `secrets.py`: conservative secret-like and placeholder-like checks for committed `.env` values

## Parser Layer

`src/parsers/` contains the currently supported input surface:

- `dotenv.py`: `.env` and `.env.example` parsing with duplicate preservation and warnings
- `python_ast.py`: `os.environ[...]` and `os.getenv(...)`
- `compose.py`: Docker Compose `${VAR}` interpolation
- `github_actions.py`: workflow expression scanning for `secrets.*` and `vars.*`

## Model Layer

`src/models.py` defines the shared data contract:

- `EnvVarDefinition`
- `EnvVarUsage`
- `EnvVarContract`
- `Finding`
- `RepoScanResult`
- `JsonEnvelope`

The JSON envelope is versioned and intended to remain a stable machine contract.

## Utilities

`src/utils/` provides deterministic helpers for:

- value normalization
- stable ordering
- repo traversal
- nearest file resolution

## Directory Map

```text
envdiff                    # Local launcher wrapper
src/analyzers/             # Comparison, scan, doctor, alias, and secret logic
src/parsers/               # Dotenv, Python AST, Compose, and GitHub Actions scanners
src/render/                # Human and JSON renderers
src/utils/                 # Ordering, normalization, and path helpers
tests/fixtures/            # Runnable example repos and file fixtures
docs/system/               # Architecture, features, operations
                           # JSON contract and finding-code references
docs/project/              # Spec and roadmap
docs/research/             # Market and comparative analysis
docs/plans/                # Implementation planning artifacts
```
