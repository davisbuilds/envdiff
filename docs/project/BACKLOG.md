# Backlog

Future-only gaps and opportunities worth revisiting. Agents can work directly
from an entry; use an issue when discussion or coordination helps. Keep one
detailed owner and reconcile affected entries when work lands. Date/source
volatile claims or label hypotheses; keep cross-repository detail with the
capability owner.

## Open

- **Typed JSON envelope** — `JsonEnvelope.Data` is `any` so structured results
  and map-like command payloads share one envelope. Typed result structs would
  remove the reflection ambiguity and the `nil`-slice-serializes-as-`null`
  hazard (currently avoided with constructors/helpers that guarantee non-nil
  slices). Largest single cleanup now that Go is unconstrained by parity.
- **CLI parsing** — the dispatcher in `internal/cli` is intentionally thin. If
  option parsing keeps growing, consider a small internal parsing abstraction
  before reaching for a third-party CLI dependency.

### Design notes (intentional, revisit only with fixtures)

- **Python scanner is regex-based** (`internal/parsers/python.go`) — line
  regexes for documented literal patterns, not full AST semantics. This is the
  spec, not a divergence; it matches inside comments/strings and misses
  multi-line calls. Add fixtures first if broader recall is ever needed.

### Lower-priority cleanup

- The matched-quote-strip idiom is duplicated in
  `internal/parsers/github_actions.go` and `internal/dotenv/parse.go` — extract a
  shared helper.
- Resolution and baseline sorting are done inline in analyzers rather than
  through `internal/order`.
