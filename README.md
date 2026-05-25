# envdiff

`envdiff` is a local-first CLI for analyzing a repository's environment contract.

It focuses on the gap between dotenv files, source code, and deployment-oriented config: what is required, what is defined, what is stale, and what looks suspicious.

## Agent Setup

New here? Paste the prompt below into your coding agent (Claude Code, Codex, etc.) and it will install, verify against the bundled fixtures, and tell you how to run it on a real repo.

```text
Set up the `envdiff` repo for me. It's a local-first CLI that analyzes a
repository's environment-variable contract (compares .env files, scans code for env
usage, flags mismatches). Python 3.11+, uv, Typer, Pydantic. It's fully local — no
network, no secrets, no env config.

Do this, in order:
1. Install deps. Ensure `uv` is installed (https://astral.sh/uv); run
   `uv sync --extra dev` from the repo root. Clone
   git@github.com:davisbuilds/envdiff.git (or the https URL) and cd in first if
   needed.
2. Verify it runs against the bundled fixtures: `./envdiff --help`,
   `uv run pytest -q`, `uv run ruff check .`, and a real scan
   `./envdiff scan tests/fixtures/repos/simple_repo --json`. All should succeed
   offline. If any fail, show me the error and stop.
3. Report back: confirm help + tests + lint + sample scan worked, and give me the
   command to run it on my own repo (e.g. `./envdiff scan <path-to-repo>` or
   `./envdiff doctor <path>`).

Don't commit anything.
```

Prefer to do it yourself? The manual steps are below.

## Current Features

- `compare` for deterministic dotenv file comparison
- `generate` for safe `.env.example` candidate generation
- `matrix` for deterministic multi-file dotenv comparison
- `scan` for repo-local env usage and definition analysis
- `doctor` for contract validation and findings
- stable JSON output for automation and agent use
- nearest `.env` / `.env.example` resolution for repo-local and monorepo layouts
- GitHub Actions workflow expression scanning for `secrets.*` and `vars.*`
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
./envdiff compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env
./envdiff generate tests/fixtures/repos/simple_repo --annotate
./envdiff generate tests/fixtures/repos/simple_repo --check
./envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env
./envdiff scan tests/fixtures/repos/workflow_repo --json
./envdiff scan tests/fixtures/repos/simple_repo --json
./envdiff doctor tests/fixtures/doctor/project --fail-on warning
```

The repo includes `./envdiff` as a local launcher, so you do not need to type
`uv run envdiff ...` during development. The fixture repos under
`tests/fixtures/` are intentionally runnable examples for local inspection.

## Project Structure

```text
envdiff/
├── docs/
│   ├── plans/
│   ├── project/
│   ├── research/
│   └── system/
├── envdiff
├── src/
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
- JSON contract reference: [docs/system/JSON_SCHEMA.md](docs/system/JSON_SCHEMA.md)
- Finding code reference: [docs/system/FINDING_CODES.md](docs/system/FINDING_CODES.md)
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
