# envdiff

Local-first CLI for analyzing a repository's environment-variable contract.
`envdiff` compares dotenv files, source-code usage, and deployment-oriented
configuration to show what is required, defined, stale, suspicious, or missing.

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

## What It Does

- Compares dotenv files deterministically with `compare`.
- Generates safe `.env.example` candidates with `generate`.
- Builds multi-file dotenv matrices with `matrix`.
- Scans repos for environment usage and definitions with `scan`.
- Validates environment contracts and findings with `doctor`.
- Emits stable JSON output for automation and agent workflows.
- Resolves nearest `.env` and `.env.example` files for repo and monorepo layouts.
- Scans GitHub Actions expressions for `secrets.*` and `vars.*`.
- Flags alias, secret-like, placeholder-like, baseline, and suppression patterns.

## Quick Start

Requirements:

- Python `3.11+`
- `uv`
- Go `1.26.4` for side-by-side Go validation

```bash
uv sync --extra dev
./envdiff --help
./envdiff scan tests/fixtures/repos/simple_repo --json
```

The repo includes `./envdiff` as a local launcher, so development commands do
not need `uv run envdiff ...`.

## Common Commands

```bash
./envdiff compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env
./envdiff generate tests/fixtures/repos/simple_repo --annotate
./envdiff generate tests/fixtures/repos/simple_repo --check
./envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env
./envdiff scan tests/fixtures/repos/workflow_repo --json
./envdiff scan tests/fixtures/repos/simple_repo --json
./envdiff doctor tests/fixtures/doctor/project --fail-on warning

uv run pytest -q
uv run ruff check .
uv run ruff format --check .
go test ./...
uv run python scripts/check_go_parity.py
```

Fixture repos under `tests/fixtures/` are intentionally runnable examples for
local inspection.

## Output Contracts

- Human output is optimized for local review.
- JSON output is deterministic and intended for automation.
- Finding codes and severity are treated as public contracts.
- Baseline snapshots and ignore suppressions are stable adoption tools.

See [docs/system/JSON_SCHEMA.md](docs/system/JSON_SCHEMA.md) and
[docs/system/FINDING_CODES.md](docs/system/FINDING_CODES.md) before changing
machine-readable output or finding semantics.

## Code Layout

```text
src/envdiff/analyzers/       repo and contract analyzers
src/envdiff/parsers/         dotenv, source, and workflow parsers
src/envdiff/render/          human and JSON renderers
src/envdiff/utils/           shared utility code
tests/                       pytest suite and fixture repos
docs/                        system, project, spec, archive, and reference docs
envdiff                      local launcher
```

## Documentation

- Agent guidance: [AGENTS.md](AGENTS.md)
- Architecture and code organization: [docs/system/ARCHITECTURE.md](docs/system/ARCHITECTURE.md)
- Current capabilities and finding surface: [docs/system/FEATURES.md](docs/system/FEATURES.md)
- JSON contract reference: [docs/system/JSON_SCHEMA.md](docs/system/JSON_SCHEMA.md)
- Finding code reference: [docs/system/FINDING_CODES.md](docs/system/FINDING_CODES.md)
- Local setup, verification, and fixture usage: [docs/system/OPERATIONS.md](docs/system/OPERATIONS.md)
- Detailed product specification: [docs/project/SPEC.md](docs/project/SPEC.md)
- Roadmap snapshot: [docs/project/ROADMAP.md](docs/project/ROADMAP.md)
- Backlog and follow-up tradeoffs: [docs/project/BACKLOG.md](docs/project/BACKLOG.md)
- Draft Go port spec: [docs/specs/2026-06-18-go-port-spec.md](docs/specs/2026-06-18-go-port-spec.md)
- Archived market and competitive analysis: [docs/archive/research/MARKET_RESEARCH.md](docs/archive/research/MARKET_RESEARCH.md)
- Archived Milestone A implementation plan: [docs/archive/plans/2026-03-17-milestone-a-usable-core-implementation.md](docs/archive/plans/2026-03-17-milestone-a-usable-core-implementation.md)

## Current Boundaries

- Repo-local analysis only.
- No shell startup file parsing.
- No env loading or injection.
- No secret manager integration.
