# envdiff

`envdiff` is a local-first CLI for analyzing a repository's environment contract.

The active design references for this implementation are:

- `envdiff-spec.md`
- `envdiff-roadmap.md`
- `envdiff-market-research.md`

Milestone A scope:

- `compare` for dotenv file comparison
- `scan` for repo-local environment usage and definition analysis
- `doctor` for deterministic contract validation

Current supported inputs:

- `.env`
- `.env.example`
- Python `os.environ[...]` and `os.getenv(...)`
- Docker Compose `${VAR}` interpolation

Example commands:

- `uv run python -m envdiff.cli compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env`
- `uv run python -m envdiff.cli scan tests/fixtures/repos/simple_repo --json`
- `uv run python -m envdiff.cli doctor tests/fixtures/doctor/project --fail-on warning`

Current limitations:

- Alias detection and secret heuristics are not part of this first implementation pass.
- Repo-local resolution is nearest `.env` / `.env.example`; broader shell semantics are intentionally out of scope.

