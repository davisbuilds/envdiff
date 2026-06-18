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
