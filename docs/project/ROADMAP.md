# Roadmap

envdiff ships in usable increments that cumulatively form a deterministic,
repo-local environment-contract analyzer. Shipped detail lives in git history
and `docs/system/`; open and unscheduled work lives in
`docs/project/BACKLOG.md`. Finished backlog items land in **Shipped** below.

## Shipped

| # | Milestone | What landed |
| :- | :-------- | :---------- |
| A | Usable core | `compare`, `scan`, `doctor` over dotenv + Python/Compose usage; repo-local resolution with nearest `.env`/`.env.example`; ignore paths; deterministic traversal; core finding codes (ENV001–ENV007) with `--fail-on` exit behavior. |
| B | Practical team tool | Alias naming-drift and secret/placeholder heuristics (ENV008–ENV009); human + JSON renderers; versioned JSON envelope; per-severity summary counts; baseline/suppression workflow for gradual CI adoption. |
| C | Broader contract platform | `matrix` multi-file consistency; `generate` with `--annotate`/`--check`/`--output`; GitHub Actions workflow scanner. |
| — | Go port | Full side-by-side Go reimplementation; Go is the default `./envdiff` launcher with Python retained as `scripts/envdiff-python`, a parity oracle for one release window; CI gates Go tests and Python/Go parity. |

## Recent hardening (2026-06)

CLI correctness, robustness, and performance pass driven by a code review of the
Go port (resolved items from the backlog):

- **Output contract** — JSON no longer HTML-escapes `<`, `>`, `&`; they appear literally in values.
- **Exit codes** — `generate --check --json` exits `2` on drift (was `0`); `--fail-on` is case-insensitive.
- **Robustness** — files are read without `bufio.Scanner`'s 64 KB line cap (new `internal/lines` reader); `generate --check` against a directory/non-file target reports drift instead of erroring.
- **Performance** — repository files are parsed in a bounded worker pool with a deterministic merge; doctor definition-name sets are memoized per file.
- **Launcher** — `./envdiff` execs a cached compiled binary (rebuilding only when sources change) instead of `go run`, cutting steady-state startup from ~45 ms to ~2 ms (see `docs/benchmarks/`).

## Open (tracked, not scheduled)

Detail in `docs/project/BACKLOG.md`:

- **Go as source of truth** — raw-UTF-8 JSON (drop `ensure_ascii`-style escaping), regenerate goldens from Go, downgrade the Python parity gate, retire the Python oracle.
- **Remaining oracle divergences** — symlink path canonicalization (`Abs` vs `resolve`), secret-length code-point counting, the usage-error exit-code contract (1 vs 2), and the intentional regex-vs-AST Python scanner.
- **Perf follow-up** — memoize the per-directory nearest-dotenv resolution walk.
- **Parser expansion (candidate)** — shell scripts, `.envrc`/direnv (repo-local subset), Pydantic `BaseSettings`, `.devcontainer`, monorepo service grouping.
- **Stretch** — pre-commit hook, editor diagnostics, alias autofix, contract export/import.

## Invariants

Determinism (same repo + flags → identical output), offline-first (no network in
core analysis), stable and documented JSON, repo-local scope (no env
loading/injection, secret distribution, or machine-global shell config),
explainable heuristics, and fail-soft handling of malformed input.

## Key decisions

- **Go is the product; Python is a transitional oracle.** Parity is a migration
  safety net, not the goal — Go-only improvements take precedence (e.g. JSON
  output), and the Python oracle is scheduled for retirement.
- **The JSON envelope, field names, and finding codes are a public contract**
  with deterministic ordering owned by `internal/order`.
