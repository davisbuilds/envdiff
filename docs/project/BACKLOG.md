# Backlog

Parking lot for implementation tradeoffs, follow-up simplifications, and product
improvement ideas that should survive across sessions. Items resolved during the
Go port and the Go-as-source-of-truth migration have been removed; see ROADMAP
"Shipped" and the migration plan for that history.

## Open

- **Secret-length code points** (`internal/normalize/value.go`) — the
  `LooksLikeSecret` length gate counts bytes via `len(value)`, so a multibyte
  value can classify differently from an equivalent ASCII one. Consider counting
  runes, and review `unicode.IsDigit` vs ASCII-digit semantics while there.
- **Resolution-walk memoization** — `resolveUsageFile` re-walks the same
  directory ancestry for every usage file; memoize the nearest
  `.env`/`.env.example` per directory. (Scan concurrency hides but does not
  eliminate the redundant walks.)
- **Typed JSON envelope** — `JsonEnvelope.Data` is `any` so structured results
  and map-like command payloads share one envelope. Typed result structs would
  remove the reflection ambiguity and the `nil`-slice-serializes-as-`null`
  hazard (currently avoided with constructors/helpers that guarantee non-nil
  slices). Largest single cleanup now that Go is unconstrained by parity.
- **CLI parsing** — the dispatcher in `internal/cli` is intentionally thin. If
  option parsing keeps growing, consider a small internal parsing abstraction
  before reaching for a third-party CLI dependency.

## Design notes (intentional, revisit only with fixtures)

- **Python scanner is regex-based** (`internal/parsers/python.go`) — line
  regexes for documented literal patterns, not full AST semantics. This is the
  spec, not a divergence; it matches inside comments/strings and misses
  multi-line calls. Add fixtures first if broader recall is ever needed.

## Lower-priority cleanup

- The matched-quote-strip idiom is duplicated in
  `internal/parsers/github_actions.go` and `internal/dotenv/parse.go` — extract a
  shared helper.
- Resolution and baseline sorting are done inline in analyzers rather than
  through `internal/order`.
