---
date: 2026-06-18
topic: go-source-of-truth-migration
stage: implementation-plan
status: complete
source: conversation
---

# Go-as-Source-of-Truth Migration Plan

## Goal

Make the Go implementation the single source of truth for `envdiff` and retire
the Python oracle. Today the repo is a side-by-side dual implementation: Python
(`src/envdiff/`) is the oracle, Go (`cmd/` + `internal/`) is the product, and a
binding apparatus (goldens generated *from Python*, `check_go_parity.py`,
`tests/test_go_parity.py`, `scripts/envdiff-python`, `pyproject.toml`/`uv.lock`,
and a parity-focused CI job) exists only to keep them in lockstep. This plan
dismantles that binding apparatus in reversible stages so the safety net is
never lost mid-flight.

Steering decision (carried from prior sessions): **Go is the product; Python is
a transitional oracle.** Parity is a migration safety net, not the goal —
Go-only improvements take precedence.

## Target end-state structure

```
envdiff/
  cmd/envdiff/         main
  internal/            all packages (unchanged)
  go.mod / go.sum
  tests/fixtures/      kept — Go already consumes these
  tests/golden/json/   kept, but Go-generated
  docs/
  .github/workflows/   Go-only CI
```

Deleted at retirement: `src/`, `pyproject.toml`, `uv.lock`, all `tests/*.py` +
`conftest.py`, `scripts/check_go_parity.py`, `scripts/envdiff-python`,
`scripts/update_go_golden.py`, ruff config/caches. The dual launcher collapses:
`./envdiff` stays as the dev rebuild-shim (or is replaced by `go install`).

Two structural inversions are the heart of the migration:

| Concern | Today | After |
| :-- | :-- | :-- |
| Golden direction | generated *from Python* CLI | generated *from Go* |
| Parity gate | hard CI gate, Python-matching | demoted to advisory → deleted with Python |
| Dead-code gate | Python `test_dead_code.py` | `staticcheck` + `deadcode` in CI |
| Lint/format | `ruff` / `ruff format` | `golangci-lint` / `gofumpt` |

## Phase 1 — Source-of-truth flip (this branch, reversible) — DONE

The authority flip only. Python stays in the tree.

Findings that shaped the implementation:

- **Go already emits raw UTF-8.** `encoding/json` does not `\uXXXX`-escape
  non-ASCII, and the earlier hardening pass already stopped HTML-escaping
  `<>&`. So no Go source change was needed for the encoding itself.
- **Go's output already matches the committed goldens byte-for-byte** for every
  existing fixture, so regenerating from Go is a no-op there. The authority flip
  is therefore safe and invisible for current cases.
- **The divergence is only observable in on-disk golden bytes**, because both
  the Go golden tests and `check_go_parity.py` compare *decoded* JSON values, so
  `é`-vs-raw-`é` is erased before comparison.

Shipped:

- `scripts/update_go_golden.py` now builds and runs the **Go binary** as the
  data source and renders with `ensure_ascii=False`. (A Go-native generator is
  deferred to phase 3, when Python is deleted.)
- New `tests/fixtures/repos/unicode_repo` + `scan-unicode-repo.json` golden, plus
  `TestScanRepositoryEmitsRawUTF8`, which pins raw UTF-8 at the **byte level**
  (the structural golden compare can't see it). Added to the parity script too.
- Docs: `docs/system/JSON_SCHEMA.md` names Go the contract source and documents
  the encoding; golden README, ROADMAP, BACKLOG updated.

**Sequencing change from the draft:** the parity gate is *kept blocking*, not
demoted to advisory. Nothing in phase 1 makes it red (decoded comparison), so it
stays valuable as a safety net through the phase-2 coverage audit. Demotion moves
to whichever follow-up first introduces a deliberate Go-only divergence (e.g. the
typed-envelope work, which changes `null`-vs-`[]`).

## Phase 2 — Coverage audit (the gate before deletion) — DONE

The ~1.6k lines of Python tests are a behavior spec. Audited each Python test
file against its Go counterpart. Most areas were already subsumed (dotenv,
python scanner, compose, github_actions, secrets, compare, matrix, generate,
models, scan, baseline, order, normalize, paths). Four genuine gaps that would
have lost coverage on Python deletion, now backfilled in Go:

- **Human rendering byte-exact output** — Go previously only `strings.Contains`
  smoke-tested human output (and had no compare/matrix human test at all). Added
  byte-exact tests for compare, scan (status join + empty `(none)`), matrix
  (reason tags, per-file presence, empty view), generate (drift/matches/write),
  and doctor (empty + severity grouping) in `internal/render/human_test.go`.
  Resolves the BACKLOG "human output is smoke-tested only" caveat.
- **Doctor behaviors** — added ENV005 template skew, the `.env.example` ENV003
  skip, the no-associated-`.env` details string, ENV007 alias **dedup** for a
  repeated missing usage, and doctor over `workflow_repo`
  (`internal/analyzers/doctor_test.go`).
- **Alias canonical-expansion match** — exact-name skip + score 0.99 +
  "Canonical token expansion matches" reason
  (`internal/analyzers/aliases_test.go`).
- **Default `.envdiffignore` discovery** — CLI auto-loads a repo-root
  `.envdiffignore` and suppresses matching findings
  (`internal/cli/cli_test.go`).

Intentionally not ported (Python-implementation-specific): `main_invokes_typer_app`
and `module_entrypoint_invokes_main` (Typer wiring), and `test_dead_code.py`
(Python static-analysis tooling — replaced by the Go lint stack in phase 3).

Exit criteria met: every behavior asserted by a Python test has a corresponding
Go assertion or a documented reason it no longer applies.

## Phase 3 — Retirement — DONE

All of the following landed:

- Delete `src/`, Python tests, `pyproject.toml`, `uv.lock`, `conftest.py`,
  `scripts/check_go_parity.py`, `scripts/envdiff-python`,
  `scripts/update_go_golden.py` (replaced by the Go generator from phase 1).
- Adopt the Go lint stack: `gofumpt` (format), `golangci-lint` (meta-linter
  bundling staticcheck/revive/errcheck/ineffassign/unused), `deadcode` for the
  cross-file dead-code gate. Add `.golangci.yml`.
- Rewrite CI to Go-only (drop the Python `lint`/`test` jobs and the uv setup
  from `go`).
- Settle the two genuine contract decisions (**decided 2026-06-18**): resolve
  symlinks in path canonicalization (`filepath.EvalSymlinks`, matching the old
  Python `resolve()`), and standardize usage errors on **exit 1** (current Go
  behavior; do not adopt Typer's exit 2). Demote/remove the parity gate as part
  of this, since the exit-code choice makes it diverge from the Python oracle.
- Update all `docs/system/` + README + AGENTS.md to drop the dual-implementation
  framing.

## Optimizations unlocked once parity is gone (separate follow-ups)

Deliberately **not** folded into phase 1 — keep the behavior flip tight and
reversible. Land these after the parity gate is advisory:

- **Typed JSON envelope** — replace `JsonEnvelope.Data any` with typed result
  structs; removes the `nil`-slice-serializes-as-`null` hazard and reflection
  ambiguity, makes the contract self-documenting.
- **Resolution-walk memoization** — `resolveUsageFile` re-walks ancestry per
  usage; memoize nearest `.env`/`.env.example` per directory.
- **Path canonicalization** — resolve symlinks once at root and cache.
- **Python scanner decision** — declare the regex scanner the spec (fast,
  zero-dep); only reach for tree-sitter-python if recall complaints surface.
- **Distribution** — `--version` from build info (`internal/version` exists),
  tagged releases (goreleaser), Homebrew tap; demote the shell launcher to dev.
- Cleanup: dedupe the quote-strip helper; route resolution/baseline ordering
  through `internal/order`.

## Invariants preserved throughout

Determinism, offline-first core, stable + documented JSON envelope and finding
codes, repo-local scope, explainable heuristics, fail-soft on malformed input.
The JSON envelope, field names, and finding codes remain a public contract with
deterministic ordering owned by `internal/order`.
