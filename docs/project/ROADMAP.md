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
| — | Go port | Full side-by-side Go reimplementation, validated against the Python oracle via a parity gate during the transition. |
| — | Go as source of truth | Go is now the sole implementation. Goldens generate from the Go binary (raw UTF-8, byte-level pinned); the Python oracle and all binding apparatus are deleted; lint moved to `golangci-lint`/`gofumpt`; CI is Go-only. Contract decisions settled: symlink-resolved paths (`EvalSymlinks`) and usage errors exit `1`. |

## Recent hardening (2026-06)

CLI correctness, robustness, and performance pass driven by a code review of the
Go port (resolved items from the backlog):

- **Output contract** — JSON no longer HTML-escapes `<`, `>`, `&`; they appear literally in values.
- **Exit codes** — `generate --check --json` exits `2` on drift (was `0`); `--fail-on` is case-insensitive.
- **Robustness** — files are read without `bufio.Scanner`'s 64 KB line cap (new `internal/lines` reader); `generate --check` against a directory/non-file target reports drift instead of erroring.
- **Performance** — repository files are parsed in a bounded worker pool with a deterministic merge; doctor definition-name sets are memoized per file.
- **Resolution-walk memoization** — nearest `.env` and `.env.example` files are
  memoized per directory during scans, avoiding repeated ancestor walks for
  usage files in the same subtree while retaining the existing resolution
  output.
- **Launcher** — `./envdiff` execs a cached compiled binary (rebuilding only when sources change) instead of `go run`, cutting steady-state startup from ~45 ms to ~2 ms (see `docs/benchmarks/`).
- **Doctor alias pruning** — the O(usages × defs) alias pass now builds a per-file `AliasIndex` (canonical tokens + inverted token index) and only compares names sharing a token, turning a scaling cliff back to ~linear (`doctor` at 2k files: 1.82 s → 4.4 ms, output unchanged).
- **Secret-length code points** — `LooksLikeSecret` now gates on
  `utf8.RuneCountInString` instead of `len(value)`, so a multibyte value
  classifies the same as an equivalent-length ASCII one; the digit count also
  switched from `unicode.IsDigit` to ASCII-only digits, matching `allDigits`
  elsewhere in the package (a non-ASCII digit code point no longer counts on
  its own).

## Open (tracked, not scheduled)

Detail in `docs/project/BACKLOG.md`:

- **Parser expansion (candidate)** — Pydantic `BaseSettings`, `.devcontainer`,
  and monorepo service grouping.
- **Stretch** — pre-commit hook, editor diagnostics, alias autofix, contract export/import.

## Invariants

Determinism (same repo + flags → identical output), offline-first (no network in
core analysis), stable and documented JSON, repo-local scope (no env
loading/injection, secret distribution, or machine-global shell config),
explainable heuristics, and fail-soft handling of malformed input.

## Key decisions

- **Go is the product.** The Python implementation was a transitional parity
  oracle for the port and has been retired; Go is the sole source of truth.
- **The JSON envelope, field names, and finding codes are a public contract**
  with deterministic ordering owned by `internal/order`. JSON is raw UTF-8;
  goldens are generated from Go.
- **Path canonicalization resolves symlinks; usage errors exit `1`** (finding-based
  exits — threshold breach, generate `--check` drift — exit `2`).
