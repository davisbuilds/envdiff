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

Behavioral divergences between the Go port and the Python oracle that current
golden fixtures / the parity gate do **not** exercise, so they pass `go test`
and `scripts/check_go_parity.py` today while still breaking the "JSON envelope +
finding codes are a public contract" guarantee on inputs no fixture contains.
Ranked roughly by severity. Each should ideally get a parity fixture before or
alongside a fix.

- **JSON encoder breaks the byte-for-byte envelope contract**
  (`internal/render/json.go`, same root cause in `internal/analyzers/baseline.go`).
  Go `json.MarshalIndent` HTML-escapes `<`, `>`, `&` (→ `<`…) and emits raw
  UTF-8; Python `json.dumps(..., sort_keys=True)` defaults to `ensure_ascii=True`
  (escapes all non-ASCII to `\uXXXX`) and does not HTML-escape. A value like
  `DATABASE_URL=postgres://h/db?a=1&b=2` or `PASSWORD=café` produces different
  bytes in each implementation; the baseline writer has the same defect on
  free-text `reason`/`details`, so Python- and Go-written baselines won't
  round-trip. Fix: `json.Encoder` with `SetEscapeHTML(false)` plus non-ASCII
  escaping to mimic `ensure_ascii`. No current fixture contains these chars.
- **`generate --check --json` returns exit 0 on drift instead of 2**
  (`internal/cli/cli.go`, `runGenerate`). The `--json` branch returns before the
  `if !checkMatches { return 2 }` block; Python raises `typer.Exit(code=2)`
  unconditionally after emitting JSON. A CI gate `envdiff generate . --check
  --json` always passes under Go. Parity `EXIT_CASES` only covers
  `generate --check` without `--json`.
- **`--fail-on` is case-sensitive in Go, case-insensitive in Python**
  (`internal/analyzers/doctor.go:ShouldFail`). Python `_should_fail` does
  `threshold.lower()`; Go switches on the exact string. `doctor . --fail-on
  ERROR` exits 0 on a clean repo under Python but exits 2 (before scanning, with
  empty stdout) under Go. New divergence not covered by the existing exit-code
  entry above.
- **Python usage detection uses line regex, not the oracle's AST**
  (`internal/parsers/python.go`). Python uses `ast.parse`. Consequences:
  commented/quoted `os.getenv("FOO")` text yields a phantom usage in Go but
  nothing in Python; a call split across lines is captured by Python and missed
  by Go; a syntactically invalid `.py` raises `SyntaxError` in Python but is
  silently regex-scanned by Go. (See the existing "Scanner boundary" note — this
  is the concrete parity cost of that intentional simplification.)
- **Path canonicalization: `filepath.Abs` vs `Path.resolve()`**
  (`internal/analyzers/scan.go`, `internal/paths/walk.go`). Go `Abs` is lexical;
  Python `resolve()` resolves symlinks. On macOS `$TMPDIR` (a symlink) or any
  symlinked component, `root_path` and every emitted `file_path`/resolution note
  diverge. Fix: `filepath.EvalSymlinks`.
- **`bufio.Scanner` line reader diverges from `str.splitlines()`**
  (`internal/dotenv/parse.go`, `internal/parsers/python.go:readLines`).
  `bufio.Scanner` splits only on `\n` and caps lines at 64 KB; `splitlines()`
  also splits on `\v`, `\f`, ` `/` `, etc. Unicode line separators
  shift `line_number`s (→ different `suppression_key`s / finding order); a line
  over 64 KB makes Go abort the whole scan with `bufio.ErrTooLong` where Python
  parses fine.
- **`LooksLikeSecret` length gate counts bytes, not characters**
  (`internal/normalize/value.go`). `len(value) < 20` is a byte count; Python
  `len(value)` is a code-point count, so a multibyte value can flip ENV008.
  `unicode.IsDigit` vs Python `str.isdigit()` also differ on superscripts/other
  digit categories.
- **Usage errors exit 1 instead of Typer's 2** (`internal/cli/cli.go`, multiple
  `return 1` paths). `compare <onefile>`, missing-path, and unknown-command exit
  1; Typer exits 2. `matrix` hardcodes 2 and is the only case the parity gate
  checks. Related to the existing exit-code entry above.
- **`CheckGeneratedExample` aborts on a directory / unreadable target**
  (`internal/analyzers/generate.go`). Go `os.ReadFile` returns the error for any
  non-`IsNotExist` failure (CLI prints `generate check failed`, exits 1); Python
  `target_path.is_file()` yields `exists=False, matches=False` and reports drift
  gracefully.

Lower-priority cleanup (all faithfully mirror the Python module layout, so not
bugs): the matched-quote-strip idiom is duplicated in
`internal/parsers/github_actions.go` and `internal/dotenv/parse.go` (extract a
shared helper); `definitionNames` is recomputed per-usage in
`internal/analyzers/doctor.go` rather than memoized by file path; resolution and
baseline sorting is done inline in analyzers instead of through `internal/order`
where the other output orderings live.
