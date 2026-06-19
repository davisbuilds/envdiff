---
date: 2026-06-19
topic: parser-coverage-expansion
stage: spec
status: in-progress
source: conversation
---

# Parser Coverage Expansion (Track B)

## Problem

`envdiff scan`/`doctor` detect environment-variable **usage** in three sources
today: Python (`os.environ`/`os.getenv`), Docker Compose interpolation, and
GitHub Actions expressions. Most real repos reference env vars from more places —
JavaScript/TypeScript, shell scripts, Dockerfiles, direnv, framework settings.
Every unseen reference is a gap in recall: `doctor` cannot flag a
referenced-but-undefined variable (ENV001/ENV002) it never saw, so its core
value is bounded by how much of the env surface it can read.

This spec expands usage detection while keeping envdiff a deterministic,
repo-local, conservative analyzer (no new runtime deps, no network, no env
loading/injection).

## Goal

Broaden the set of files whose env-var references envdiff understands, so `scan`
contracts and `doctor` findings cover the env surface a typical polyglot repo
actually has — without raising the false-positive rate or changing the JSON
contract's shape.

## Design constraints & decisions

- **Reuse the existing parser shape.** Each new parser is a function in
  `internal/parsers/` that takes a path, uses `lines.Read`, applies documented
  line regexes, and returns `model.UsageScanResult{Usages, Warnings}` ordered by
  `order.Usages`. Dispatch is a new `case` in `scanFile` (scan.go). This is the
  established pattern (`python.go`, `compose.go`, `github_actions.go`).
- **Conservative recall over cleverness.** Line-oriented regexes for documented
  literal patterns only — same boundary as the Python scanner. Dynamic/computed
  names (`process.env[key]`, `os.environ[f"{x}"]`) are intentionally skipped. Add
  fixtures before broadening.
- **Requiredness reflects language semantics**, mapped to the existing
  vocabulary (`required` / `optional` / `optional_with_default`):
  - A read that *throws/fails* when the var is absent → `required`.
  - A read that returns null/undefined/empty when absent → `optional`.
  - A read with an inline fallback (`|| x`, `?? x`, `:-x`, `ENV K=default`) →
    `optional_with_default` with the captured default.
- **`source_type` is a new stable enum value per language** (`javascript`,
  `shell`, `dockerfile`, …); `usage_kind` describes the access form
  (`process.env`, `shell_var`, `dockerfile_arg`, …). These appear in the public
  JSON contract, so values are chosen once and kept stable.
- **No new definition semantics in the first slice.** Shell `export`/Dockerfile
  `ENV`/`ARG` both *define* and *use* vars, which entangles the doctor model
  (a locally-defined var referenced in the same script must not be flagged
  missing). That is real work and is deferred to phase B2 so it can be designed
  deliberately; the first slice ships a pure usage parser with no model change.
- **Determinism preserved.** Output ordering stays owned by `internal/order`;
  goldens are regenerated from Go.

## Phasing

Phases are sequenced by value-to-risk. Each is independently shippable.

### Phase B1 — JavaScript/TypeScript `process.env` (DONE)

The cleanest, highest-value first parser: it maps onto the existing usage model
exactly like the Python parser, with **no define/use entanglement** (assigning
`process.env.X` is rare and out of scope), so it adds recall with zero model
change. Directly relevant to the Next.js-heavy side of the portfolio.

- **Files:** `.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs`.
- **Detect:**
  - `process.env.NAME` (dot access) and `process.env['NAME']` /
    `process.env["NAME"]` (bracket access), where `NAME` is
    `[A-Za-z_][A-Za-z0-9_]*`.
  - Inline default: `process.env.NAME || <lit>` and `process.env.NAME ?? <lit>`,
    capturing a quoted-string or bare (numeric/identifier) literal as the
    default value.
- **Requiredness:** a bare read is `optional` (JS returns `undefined`, it does
  not throw); a read with `||`/`??` fallback is `optional_with_default`. There is
  no `required` form in plain JS, which keeps findings conservative
  (missing → ENV002 warning, not ENV001 error).
- **`source_type`:** `javascript` (covers TS — the access pattern is identical).
  **`usage_kind`:** `process.env`.
- **Out of scope (B1):** destructuring (`const { FOO } = process.env`), computed
  keys, `import.meta.env` (Vite), and `.env` *loading* libraries. Noted as B3
  candidates.

### Phase B2 — Shell + Dockerfile (define + use)

Introduces local-definition awareness so a var defined and used in the same
scope is not a false "missing". Scope/decision work:

- Shell: `export NAME=…` / `NAME=…` (definition), `$NAME` / `${NAME}` /
  `${NAME:-default}` / `${NAME:=default}` (usage). Files: `*.sh`, `*.bash`, and
  shebang detection for extensionless scripts (deferred sub-item).
- Dockerfile: `ARG NAME[=default]`, `ENV NAME=value` / `ENV NAME value`
  (definitions), `${NAME}` / `$NAME` (usage). File: `Dockerfile`, `*.Dockerfile`.
- **Open decision (resolve in B2 spec update):** do shell/Dockerfile-local
  definitions satisfy a contract the way a `.env` definition does, or are they a
  separate scope that only suppresses same-file "missing" findings? Likely the
  latter, to avoid cross-scope confusion in the doctor.

### Phase B3 — Framework & remaining sources (later)

direnv `.envrc`, Pydantic `BaseSettings` / framework settings, Compose
`env_file`, Vite `import.meta.env`, JS destructuring, monorepo service grouping.
Each gets fixtures and a parser; spec'd when picked up.

## Per-phase implementation checklist (applies to every parser)

1. New `internal/parsers/<lang>.go` returning `model.UsageScanResult`.
2. Dispatch `case` in `scanFile` (and a filename/ext predicate if needed).
3. Fixtures under `tests/fixtures/<lang>/` (a representative file) and, for the
   scan integration, a fixture repo or an addition to an existing one.
4. Focused parser test (`internal/parsers/<lang>_test.go`) covering each
   requiredness path and the "ignored dynamic names" case.
5. A scan integration assertion (new contract names appear) and golden
   regeneration (`ENVDIFF_UPDATE_GOLDENS=1 go test ./...`).
6. Update `docs/system/FEATURES.md` (supported inputs) and `ARCHITECTURE.md`
   (parser list); note the new `source_type`/`usage_kind` values in
   `docs/system/JSON_SCHEMA.md` if it enumerates them.

## Non-goals

- No `.env`-file loading, env injection, or secret resolution.
- No full language parsing/AST; line regexes only.
- No change to the JSON envelope shape, finding codes, or ordering semantics.
- No network or new runtime dependencies.

## Success criteria

- `scan` over a JS/TS fixture surfaces `process.env` references as contracts with
  correct requiredness and defaults; `doctor` flags undefined ones as ENV002.
- Parser + integration tests pass; goldens regenerate deterministically; the full
  Go gate (`go test ./...`, `golangci-lint run ./...`) is green.
- No new false positives on the existing fixtures.
