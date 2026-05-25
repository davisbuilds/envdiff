# AGENTS.md

`envdiff` is a deterministic CLI that scans repos for environment variable contracts,
compares `.env` files, and flags mismatches. Local-first; no network in core analysis.

## Documentation Map

- `docs/system/ARCHITECTURE.md` — high-level flow, CLI/analyzer/parser/model/utils layers, directory map.
- `docs/system/FEATURES.md` — supported inputs, command table, finding codes (ENV001–009), heuristics, output modes, deferred features.
- `docs/system/OPERATIONS.md` — setup, runnable fixture repos, exit codes, constraints, implementation status.
- `docs/system/JSON_SCHEMA.md` — JSON envelope contract.
- `docs/system/FINDING_CODES.md` — finding-code reference.
- `docs/project/SPEC.md` — problem framing and in/out scope.
- `docs/project/ROADMAP.md` — shipped highlights and open items.

## Command Quickstart

```bash
uv sync --extra dev                  # install deps + dev tools
./envdiff --help                     # list all commands (or: uv run envdiff <cmd>)
uv run pytest -q                     # tests
uv run ruff check .                  # lint
```

Commands: `compare`, `scan`, `matrix`, `doctor`, `generate` — each has a human path and a `--json` path. Entry point: `src.cli:main` (Typer).

## Project Boundaries

`envdiff` is a deterministic, repo-local env contract analyzer — keep it that way.

- Don't expand it into an env loader, injector, or secret-distribution tool.
- Keep analysis repo-local; user-shell startup files are out of scope.
- Prefer conservative heuristics over aggressive recall.
- **Treat the JSON envelope, field names, and finding codes as a public contract**, and keep finding/contract ordering deterministic. When CLI behavior changes, update `docs/system/` in the same change.

## Testing

- **Pre-push**: `uv run ruff check .` and `uv run pytest -q`.
- **TDD**: red/green for new features and major changes.
- Favor behavior-oriented tests over implementation detail; use real fixture repos under `tests/fixtures/` instead of mocks.
- For parser work add focused parser tests plus a repo-scan integration test; for CLI changes update `tests/test_cli_smoke.py`.
- **Dead-code gate** (`tests/test_dead_code.py`): static checks for unused public symbols, orphaned modules, and unreachable code. It owns cross-file dead code; ruff `F`/`ERA` own within-file unused imports/locals and commented-out code. When a symbol/module is intentionally unreferenced (external API, framework-invoked), add it to `SYMBOL_EXCEPTIONS`/`MODULE_EXCEPTIONS` with a reason rather than silencing the test.

## Conventions Enforced Elsewhere

Ruff handles modern-Python style (import sorting, `from __future__ import annotations`, upgrade rules) — fix violations the linter flags rather than restating them here.

## Working Agreement

- **Push back before building.** If a request is incoherent or self-contradictory, or a spec/plan is vague or skips key decisions, stop and interview me — ask clarifying questions and confirm intent before writing code or changing files. Don't guess at scope or comply silently. (Clear, well-scoped requests don't need this.)
- **Keep docs current.** After a significant change, PR, or completed spec/plan, update any now-stale reference docs under `docs/system/` (and `docs/project/ROADMAP.md`) so they match shipped behavior. Skip this for trivial changes.
