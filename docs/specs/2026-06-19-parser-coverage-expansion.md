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

### Phase B2 — Shell + Dockerfile (define + use) (DONE)

Introduces local-definition awareness so a var defined and used in the same file
is not a false "missing".

- Shell: `export NAME=…` / `NAME=…` (local definition), `$NAME` / `${NAME}` /
  `${NAME:-default}` / `${NAME:=default}` / `${NAME:?…}` (usage). Files: `*.sh`,
  `*.bash` (shebang detection for extensionless scripts deferred to B3).
- Dockerfile: `ARG NAME[=default]`, `ENV NAME=value` / `ENV NAME value`
  (local definitions), `${NAME}` / `$NAME` (usage). Files: `Dockerfile`,
  `Dockerfile.*`, `*.Dockerfile`.
- **Decision (2026-06-19): separate scope.** Local `export`/`ENV`/`ARG`
  definitions do **not** satisfy `.env` contracts elsewhere in the repo and are
  not emitted into the scan's `Definitions`/contracts. They only suppress
  "missing" findings for usages of the same name **in the same file**.
  Implementation stays entirely in the parser: detect local definitions, then
  emit usages only for referenced names not locally defined in that file. No
  doctor/model/contract change. A var referenced in file B but defined only in
  file A is still flagged (resolves against the nearest `.env`, as today).
- **Requiredness (follows the Compose precedent):** bare `$NAME`/`${NAME}` and
  `${NAME:?…}` → `required`; `${NAME:-d}`/`${NAME:=d}` → `optional_with_default`
  with the captured default.
- **`source_type`:** `shell`, `dockerfile`. **`usage_kind`:** `shell_var`,
  `dockerfile_var`.

### Phase B3a — Cheap recall wins (DONE)

Three additions that fit the established line-regex parser pattern with little
new modeling — bundled into one PR:

- **Vite `import.meta.env.NAME`** — another access form in `javascript.go`,
  identical model to `process.env` (optional; `|| / ??` → optional_with_default).
  `usage_kind` `import.meta.env`.
- **direnv `.envrc`** — direnv files are shell syntax, so route `.envrc` (and
  `.envrc.local`) to the existing shell scanner via a dispatch addition. Local
  `export`/assignment suppression already applies. No new parser.
- **JS/TS destructuring** — `const { A, B } = process.env` (single line only;
  multi-line strains the line scanner and stays out of scope). Each destructured
  name is an `optional` `process.env` usage. Renames (`{ A: b }`) capture the
  source key `A`.

### Phase B3b — Bigger, à-la-carte (deferred; spec each when picked up)

These do **not** fit the line-regex usage-parser mold and each carries its own
design decision, so they are not bundled and are best done individually (likely
after Track A — distribution):

- **Compose `env_file:`** — a *resolution* feature, not a usage parser: it
  declares which env file(s) a service loads. Touches resolution logic and the
  question of whether a compose-referenced env file is a definition source.
- **Pydantic `BaseSettings` / framework settings** — *declaration* modeling:
  class fields (with `env_prefix`, `Field(alias=…)`, defaults) declare env vars
  across multiple lines. Different from both `.env` definitions and simple
  usages; needs design and multi-line handling.
- **Monorepo service grouping** — a scan/doctor *output/reporting* feature
  (group contracts per service subtree), not a parser. Resolution already finds
  the nearest `.env` per file; grouping is presentation.

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
