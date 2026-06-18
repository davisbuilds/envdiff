# Backlog

Parking lot for implementation tradeoffs, follow-up simplifications, and product
improvement ideas that should survive across sessions.

## Go Port Follow-Ups

- **Python legacy ownership:** Python is retained as a legacy parity oracle for
  one release window after the Go launcher cutover. Decide later whether to move
  it to a legacy path or remove it after confidence in the Go implementation is
  sufficient.
- **Toolchain provenance:** Go was installed via Homebrew during the side-by-side
  port work. Current local version is `go1.26.4 darwin/arm64`, and `go.mod`
  reflects `go 1.26.4`.
- **Module path choice:** The Go module path is `github.com/davisbuilds/envdiff`,
  inferred from the README clone URL. Revisit only if packaging or repo ownership
  changes.
- **Task sequencing note:** Task 6 required CLI edits even though the spec file
  list did not mention them, because the verification command used
  `go run ./cmd/envdiff scan ... --json`. Future specs should keep file lists and
  verification commands aligned.
- **Scanner boundary:** The Go Python scanner is intentionally line-oriented and
  regex-based for documented literal patterns only. Add fixtures first if future
  parity needs broader Python AST semantics.
- **JSON data typing:** `JsonEnvelope.Data` is `any` so structured scan output and
  map-like command payloads can both serialize through one envelope. Consider
  typed command result structs once parity is complete.
- **Nil slice risk:** Go `nil` slices serialize as `null`, while schema version `1`
  often expects `[]`. Prefer constructors/helpers or typed result structs that
  guarantee non-nil slices.
- **CLI parsing:** The local dispatcher is intentionally thin. If option parsing
  continues to grow after parity, consider a small internal parsing abstraction
  before adding a third-party CLI dependency.
- **Human rendering:** Human output is smoke-tested for key user-facing strings,
  not byte-for-byte parity. Polish should stay behind JSON parity until cutover.
- **Exit-code parity:** The spec says command usage failures exit `1`, but the
  current Python/Typer oracle exits `2` for some validation failures such as
  single-file `matrix` and invalid `doctor --fail-on`. The Go port currently
  matches Python for parity; decide later whether to document this as the public
  contract or intentionally change both implementations.

## Go Port Parity Divergences (code review 2026-06-18)

Open divergences between the Go port and the Python oracle that current fixtures
do not exercise. Resolved items shipped in the CLI hardening pass — see ROADMAP
"Recent hardening".

- **JSON non-ASCII handling** — Go emits raw UTF-8 where the Python oracle
  escapes to `\uXXXX` (`ensure_ascii`). Deliberately deferred to the
  Go-as-source-of-truth branch (regenerate goldens from Go, relax the parity
  gate); no fixture exercises non-ASCII today.
- **Path canonicalization** (`internal/analyzers/scan.go`,
  `internal/paths/walk.go`) — Go `filepath.Abs` is lexical; Python
  `Path.resolve()` resolves symlinks, so on a symlinked component (e.g. macOS
  `$TMPDIR`) `root_path` and every emitted path diverge. Fix:
  `filepath.EvalSymlinks`.
- **`LooksLikeSecret` length gate** (`internal/normalize/value.go`) —
  `len(value)` counts bytes; Python counts code points, so a multibyte value can
  flip ENV008. `unicode.IsDigit` vs `str.isdigit()` also differ on
  superscripts/other digit categories.
- **Usage-error exit codes** — wrong-arity / missing-path / unknown-command exit
  `1` in Go but `2` in Python/Typer. A contract *decision* (see "Exit-code
  parity" above), not yet resolved.
- **Python regex vs AST scanner** (`internal/parsers/python.go`) — the line
  regex matches inside comments/strings and misses multi-line calls; intentional
  (see "Scanner boundary"). Revisit only with fixtures first.
- **Resolution-walk memoization** — `resolveUsageFile` re-walks the same
  directory ancestry for every usage file; memoize the nearest
  `.env`/`.env.example` per directory. (Found while parallelizing the scan;
  concurrency hides but does not eliminate it.)

Lower-priority cleanup: the matched-quote-strip idiom is duplicated in
`internal/parsers/github_actions.go` and `internal/dotenv/parse.go` (extract a
shared helper); resolution and baseline sorting are done inline in analyzers
rather than through `internal/order`. (The per-usage `definitionNames` rebuild
and per-parser `readLines` duplication were resolved in the hardening pass.)
