# envdiff

`envdiff` is a local-first CLI for analyzing a repository's environment contract.

It focuses on the gap between dotenv files, source code, and deployment-oriented config: what is required, what is defined, what is stale, and what looks suspicious.

## Current Features

- `compare` for deterministic dotenv file comparison
- `generate` for safe `.env.example` candidate generation
- `matrix` for deterministic multi-file dotenv comparison
- `scan` for repo-local env usage and definition analysis
- `doctor` for contract validation and findings
- stable JSON output for automation and agent use
- nearest `.env` / `.env.example` resolution for repo-local and monorepo layouts
- conservative alias, secret-like, and placeholder-like heuristics
- baseline snapshots and ignore-based suppression for incremental adoption

## Tech Stack

- Python 3.11+
- `uv` for dependency and environment management
- Typer for the CLI
- Pydantic for shared models
- Rich-compatible human rendering plus deterministic JSON output

## Setup

```bash
uv sync --extra dev
```

## Usage

```bash
uv run python -m envdiff.cli compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env
uv run python -m envdiff.cli generate tests/fixtures/repos/simple_repo --annotate
uv run python -m envdiff.cli matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env
uv run python -m envdiff.cli scan tests/fixtures/repos/simple_repo --json
uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --fail-on warning
```

The fixture repos under `tests/fixtures/` are intentionally runnable examples for local inspection.

## Project Structure

```text
envdiff/
├── docs/
│   ├── plans/
│   ├── project/
│   ├── research/
│   └── system/
├── envdiff/
│   ├── analyzers/
│   ├── parsers/
│   ├── render/
│   └── utils/
├── tests/
└── pyproject.toml
```

## Documentation

- Architecture and code organization: [docs/system/ARCHITECTURE.md](docs/system/ARCHITECTURE.md)
- Current capabilities and finding surface: [docs/system/FEATURES.md](docs/system/FEATURES.md)
- Local setup, verification, and fixture usage: [docs/system/OPERATIONS.md](docs/system/OPERATIONS.md)
- Detailed product specification: [docs/project/SPEC.md](docs/project/SPEC.md)
- Roadmap snapshot: [docs/project/ROADMAP.md](docs/project/ROADMAP.md)
- Market and competitive analysis: [docs/research/MARKET_RESEARCH.md](docs/research/MARKET_RESEARCH.md)
- Active implementation plan: [docs/plans/2026-03-17-milestone-a-usable-core-implementation.md](docs/plans/2026-03-17-milestone-a-usable-core-implementation.md)

## Current Boundaries

- repo-local analysis only
- no shell startup file parsing
- no env loading or injection
- no secret manager integration
