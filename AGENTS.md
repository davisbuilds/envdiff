# AGENTS.md

`envdiff` is a deterministic CLI that scans repos for environment variable contracts,
compares `.env` files, and flags mismatches. Local-first; no network in core analysis.

## Documentation Map

- `docs/system/ARCHITECTURE.md` — high-level flow, CLI/analyzer/parser/model/utils layers, directory map.
- `docs/system/FEATURES.md` — supported inputs, command table, finding codes (ENV001–009), heuristics, output modes, deferred features.
- `docs/system/OPERATIONS.md` — setup, runnable fixture repos, exit codes, constraints, implementation status.
- `docs/system/JSON_SCHEMA.md` — JSON envelope contract.
- `docs/system/FINDING_CODES.md` — finding-code reference.
- `docs/project/SPEC.md` — problem framing and in/out scope.
- `docs/project/ROADMAP.md` — shipped highlights and open items.
- `docs/project/BACKLOG.md` — tradeoffs and follow-up simplification backlog.
- `docs/project/GIT_HISTORY_POLICY.md` — merge settings, branch protection, and history conventions.

## Command Quickstart

```bash
./envdiff --help                     # list all commands (Go launcher; builds/caches bin/envdiff)
go build -o bin/envdiff ./cmd/envdiff   # build the binary directly
go test ./...                        # tests (also validates committed goldens)
golangci-lint run ./...              # lint
ENVDIFF_UPDATE_GOLDENS=1 go test ./...  # regenerate JSON goldens from Go
```

Commands: `compare`, `scan`, `matrix`, `doctor`, `generate` — each has a human path and a `--json` path. Entry point: `cmd/envdiff` (Go); the source lives under `internal/`. envdiff is Go-only — the original Python implementation was a transitional oracle and has been retired (see `docs/project/ROADMAP.md` → "Go as source of truth").

## Project Boundaries

`envdiff` is a deterministic, repo-local env contract analyzer — keep it that way.

- Don't expand it into an env loader, injector, or secret-distribution tool.
- Keep analysis repo-local; user-shell startup files are out of scope.
- Prefer conservative heuristics over aggressive recall.
- **Treat the JSON envelope, field names, and finding codes as a public contract**, and keep finding/contract ordering deterministic. When CLI behavior changes, update `docs/system/` in the same change.

## Testing

- **Pre-push** (matches CI `.github/workflows/ci.yml`): `go vet ./...`, `go test ./...`, `golangci-lint run ./...`, and confirm goldens are current (`ENVDIFF_UPDATE_GOLDENS=1 go test ./... && git diff --exit-code tests/golden`).
- **TDD**: red/green for new features and major changes.
- Favor behavior-oriented tests over implementation detail; use real fixture repos under `tests/fixtures/` instead of mocks.
- For parser work add focused parser tests (`internal/parsers`, `internal/dotenv`) plus a repo-scan integration test (`internal/analyzers/scan_test.go`); for CLI changes update `internal/cli/cli_test.go`.
- **JSON goldens** (`tests/golden/json/`) are the rendered-output contract, generated from Go. Each golden has exactly one writer (an analyzer-level test via `testutil.AssertGoldenJSON`); CLI/render tests are pure consumers. `TestScanRepositoryEmitsRawUTF8` pins raw-UTF-8 encoding at the byte level.

## Conventions Enforced Elsewhere

`golangci-lint` (config in `.golangci.yml`) is the lint stack: `gofumpt` formatting, plus `unused`/`staticcheck` (which own cross-file dead code), `errcheck`, `govet`, `ineffassign`, and a few style linters. Fix what the linter flags rather than restating rules here.

## Working Agreement

- **Push back before building.** If a request is incoherent or self-contradictory, or a spec/plan is vague or skips key decisions, stop and interview me — ask clarifying questions and confirm intent before writing code or changing files. Don't guess at scope or comply silently. (Clear, well-scoped requests don't need this.)
- **Keep docs current.** After a significant change, PR, or completed spec/plan, update any now-stale reference docs under `docs/system/` (and `docs/project/ROADMAP.md`) so they match shipped behavior. Skip this for trivial changes.
- **Commit logically.** Commit completed work in coherent chunks as you proceed. Push only when explicitly asked.
- **Re-ground after compaction.** A compaction summary loses precise paths, context, and verification state — before continuing, re-read this project's `AGENTS.md`, its reference docs, and recent commits.
