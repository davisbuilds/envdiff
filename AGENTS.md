# AI Agent Guide

Guidance for coding agents working in this repository.

Deterministic CLI that scans repos for environment variable contracts, compares `.env` files, and flags mismatches.

## Project Structure

```text
envdiff/
├── envdiff                 # Repo-local launcher wrapper
├── src/                    # Application source code
│   ├── analyzers/          # Contract analysis, compare, doctor, matrix, generate, baselines
│   ├── parsers/            # Dotenv, Python AST, Compose, and GitHub Actions scanners
│   ├── render/             # Human and JSON renderers
│   ├── utils/              # Deterministic ordering, normalization, and path helpers
│   ├── cli.py              # Typer CLI entry point
│   └── models.py           # Shared Pydantic data models and JSON envelope
├── tests/                  # Pytest suite
│   └── fixtures/           # Runnable example repos and parser fixtures
├── docs/
│   ├── system/             # Architecture, features, operations, JSON schema, finding codes
│   ├── project/            # Spec and roadmap
│   ├── research/           # Market research
│   └── plans/              # Implementation plans
└── pyproject.toml
```

## Key Files Reference

| Purpose | Location |
|---------|----------|
| CLI entry point (Typer) | `src/cli.py` |
| Shared data models and JSON envelope | `src/models.py` |
| Repo scanner | `src/analyzers/scan.py` |
| Environment comparison | `src/analyzers/compare.py` |
| Health check / diagnostics | `src/analyzers/doctor.py` |
| Config generation | `src/analyzers/generate.py` |
| Cross-file matrix | `src/analyzers/matrix.py` |
| Baseline analysis | `src/analyzers/baseline.py` |
| Secrets detection | `src/analyzers/secrets.py` |
| Dotenv parser | `src/parsers/dotenv.py` |
| Python AST scanner | `src/parsers/python_ast.py` |
| Docker Compose parser | `src/parsers/compose.py` |
| GitHub Actions parser | `src/parsers/github_actions.py` |
| Human-readable output | `src/render/human.py` |
| JSON output | `src/render/json.py` |
| CLI smoke tests | `tests/test_cli_smoke.py` |
| Test fixtures (example repos) | `tests/fixtures/` |

## Key Commands

The project uses `uv` for dependency management.

- **Run CLI**: `./envdiff <command>` (or `uv run envdiff <command>`)
- **Run Tests**: `uv run pytest -q`
- **Lint**: `uv run ruff check .`
- **Install**: `uv sync --extra dev`

## Project Boundaries

`envdiff` is a deterministic, repo-local env contract analyzer.

- Do not expand it into an env loader, injector, or secret distribution tool.
- Keep analysis repo-local; user-shell startup files are out of scope.
- Prefer conservative heuristics over aggressive recall.
- Preserve stable JSON output and deterministic ordering across commands.

## Development Patterns

- **CLI entrypoint**: `src.cli:main` via Typer.
- **Shared models**: `src.models` defines contracts, findings, envelopes, and baselines.
- **Repo scan pipeline**: `src.analyzers.scan.scan_repository()`.
- **Validation pipeline**: `src.analyzers.doctor.doctor_repository()`.
- **Generation flow**: `src.analyzers.generate`.
- **Comparison flows**: `src.analyzers.compare` and `src.analyzers.matrix`.
- **Human output**: `src.render.human`.
- **Machine output**: `src.render.json`.

## Coding Conventions

Ruff is configured with import sorting and modern-Python upgrade rules.

- Use `from __future__ import annotations` in Python modules.
- Keep ordering deterministic for findings, contracts, usages, and definitions.
- Prefer explicit small helpers over broad abstraction layers.
- Treat JSON field names and finding codes as public contract.
- When changing CLI behavior, update the docs in `docs/system/` at the same time.

## Testing

**Pre-push check**: Before pushing updates to the remote, run `uv run ruff check .` and `uv run pytest -q`.

**TDD**: Use red/green TDD for new features and major changes.

**Key patterns**:
- Favor behavior-oriented tests over implementation-detail tests.
- Use real fixture repos under `tests/fixtures/` instead of mocks where possible.
- For parser work, add focused parser tests and at least one repo-scan integration test.
- For CLI changes, add or update smoke coverage in `tests/test_cli_smoke.py`.

## CLI Overview

Commands: `compare`, `scan`, `matrix`, `doctor`, `generate`.

Notable workflows:
- `./envdiff doctor . --fail-on warning`
- `./envdiff generate . --check`
- `./envdiff matrix path/to/a.env path/to/b.env`
- `./envdiff scan . --json`

See `docs/system/FEATURES.md`, `docs/system/JSON_SCHEMA.md`, and `docs/system/FINDING_CODES.md` for the public command and output contract.
