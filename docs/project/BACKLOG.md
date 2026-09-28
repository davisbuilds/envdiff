# Backlog

Future-only gaps and opportunities worth revisiting. Capture recurring friction,
meaningful risk or cost, unresolved decisions, or concrete revisit triggers.
Fix simple, quick, or blocking issues inline when within the active task's scope.

## Conventions

- **Entry:** state **What** and **Why or evidence**. Add **Next** (a useful first
  action) or **Revisit when** (a concrete gate) where helpful; no fixed template
  is required.
- **Evidence:** date and source volatile claims. Support causal or performance
  claims with measurements, or label them **hypothesis, unmeasured**.
- **Delegation:** agents can execute entries directly. Recording a candidate does
  not expand the active task or select a roadmap priority. Use an issue when
  persistent discussion or coordination helps; no mandatory graduation step.
- **Ownership:** keep cross-repository work with the capability-owning repository.
  If an issue owns the details, retain only a useful linked summary here; avoid
  parallel checklists. Keep private evidence out of public entries and issues.
- **Closure:** reconcile affected entries as work lands. Remove resolved concerns,
  retain unresolved remainders, and preserve durable rationale in its owning
  reference. Roadmap records selected direction; Git and PRs hold routine shipped
  history. Revisit the broader list during prioritization or when stale entries
  impede work.

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
