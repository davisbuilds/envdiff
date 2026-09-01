# Backlog

Future-only design gaps, tech debt, and better ways to do a thing noticed during normal
execution. Fix simple, quick, or blocking issues inline; capture only durable follow-ups
worth revisiting cold. Add an item only when it cannot be fixed inline and represents
recurring friction, meaningful risk or cost, an unresolved decision, or a concrete
trigger. Resolved Go-port history lives in ROADMAP "Shipped" and the migration plan,
not here.

This repository is the canonical owner for its follow-ups; cross-repository work belongs
with the repository that owns the capability, with links from affected repositories only
when useful. Date and source volatile external or runtime claims, or label them a
hypothesis.

Entries should record **What** / **Why or evidence** / optional **Next** / optional
**Revisit when**. Use **Next** for the smallest action that makes an item actionable and
**Revisit when** only for an intentional external or measurable gate.

Review this file after a significant shipped slice or at least quarterly: confirm each
item is still open, refresh dated evidence, promote selected work to a plan, convert it
to a trigger, or move completed decisions and work to the Roadmap or decision history.

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
