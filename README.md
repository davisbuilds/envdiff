# envdiff

Local-first CLI for analyzing a repository's environment-variable contract.
`envdiff` compares dotenv files, source-code usage, and deployment-oriented
configuration to show what is required, defined, stale, suspicious, or missing.

## Agent Setup

New here? Paste the prompt below into your coding agent (Claude Code, Codex, etc.) and it will install, verify against the bundled fixtures, and tell you how to run it on a real repo.

```text
Set up the `envdiff` repo for me. It's a local-first Go CLI that analyzes a
repository's environment-variable contract (compares .env files, scans code for env
usage, flags mismatches). Go 1.26.4. It's fully local — no network, no secrets,
no env config.

Do this, in order:
1. Install deps. Ensure Go 1.26.4 is on PATH. Clone
   git@github.com:davisbuilds/envdiff.git (or the https URL) and cd in first if
   needed.
2. Verify it runs against the bundled fixtures: `./envdiff --help`,
   `go test ./...`, and a real scan
   `./envdiff scan tests/fixtures/repos/simple_repo --json`. All should succeed
   offline. If any fail, show me the error and stop.
3. Report back: confirm help + tests + sample scan worked, and give me the
   command to run it on my own repo (e.g. `./envdiff scan <path-to-repo>` or
   `./envdiff doctor <path>`).

Don't commit anything.
```

Prefer to do it yourself? The manual steps are below.

## What It Does

- Compares dotenv files deterministically with `compare`.
- Generates safe `.env.example` candidates with `generate`.
- Builds multi-file dotenv matrices with `matrix`.
- Scans repos for environment usage and definitions with `scan`.
- Validates environment contracts and findings with `doctor`.
- Emits stable JSON output for automation and agent workflows.
- Resolves nearest `.env` and `.env.example` files for repo and monorepo layouts.
- Scans GitHub Actions expressions for `secrets.*` and `vars.*`.
- Flags alias, secret-like, placeholder-like, baseline, and suppression patterns.

## Quick Start

Requirements:

- Go `1.26.4`

```bash
./envdiff --help
./envdiff scan tests/fixtures/repos/simple_repo --json
```

`./envdiff` is the local launcher: it builds and caches `bin/envdiff`, rebuilding
only when sources change. You can also `go build -o bin/envdiff ./cmd/envdiff`
and run the binary directly.

## Common Commands

```bash
./envdiff compare tests/fixtures/compare/left.env tests/fixtures/compare/right.env
./envdiff generate tests/fixtures/repos/simple_repo --annotate
./envdiff generate tests/fixtures/repos/simple_repo --check
./envdiff matrix tests/fixtures/matrix/a.env tests/fixtures/matrix/b.env tests/fixtures/matrix/c.env
./envdiff scan tests/fixtures/repos/workflow_repo --json
./envdiff scan tests/fixtures/repos/simple_repo --json
./envdiff doctor tests/fixtures/doctor/project --fail-on warning

go test ./...
golangci-lint run ./...
ENVDIFF_UPDATE_GOLDENS=1 go test ./...   # regenerate JSON goldens from Go
```

Fixture repos under `tests/fixtures/` are intentionally runnable examples for
local inspection.

## Output Contracts

- Human output is optimized for local review.
- JSON output is deterministic and intended for automation.
- Finding codes and severity are treated as public contracts.
- Baseline snapshots and ignore suppressions are stable adoption tools.

See [docs/system/JSON_SCHEMA.md](docs/system/JSON_SCHEMA.md) and
[docs/system/FINDING_CODES.md](docs/system/FINDING_CODES.md) before changing
machine-readable output or finding semantics.

## Code Layout

```text
cmd/envdiff/                 CLI entrypoint
internal/analyzers/          repo and contract analyzers
internal/parsers/            dotenv, source, and workflow parsers
internal/dotenv/             dotenv parsing
internal/render/             human and JSON renderers
internal/model/              data model + JSON envelope
internal/order/              deterministic ordering
internal/paths/, lines/, normalize/, version/   shared packages
tests/fixtures/              runnable fixture repos
tests/golden/json/           rendered-output goldens (generated from Go)
docs/                        system, project, and reference docs
envdiff                      local launcher (builds/caches bin/envdiff)
```

## Documentation

- Agent guidance: [AGENTS.md](AGENTS.md)
- Architecture and code organization: [docs/system/ARCHITECTURE.md](docs/system/ARCHITECTURE.md)
- Current capabilities and finding surface: [docs/system/FEATURES.md](docs/system/FEATURES.md)
- JSON contract reference: [docs/system/JSON_SCHEMA.md](docs/system/JSON_SCHEMA.md)
- Finding code reference: [docs/system/FINDING_CODES.md](docs/system/FINDING_CODES.md)
- Local setup, verification, and fixture usage: [docs/system/OPERATIONS.md](docs/system/OPERATIONS.md)
- Detailed product specification: [docs/project/SPEC.md](docs/project/SPEC.md)
- Roadmap snapshot: [docs/project/ROADMAP.md](docs/project/ROADMAP.md)
- Backlog and follow-up tradeoffs: [docs/project/BACKLOG.md](docs/project/BACKLOG.md)
- Git history and branch policy: [docs/project/GIT_HISTORY_POLICY.md](docs/project/GIT_HISTORY_POLICY.md)
- Go vs Python benchmark (historical, from the port): [docs/benchmarks/2026-06-18-go-vs-python.md](docs/benchmarks/2026-06-18-go-vs-python.md)

## Current Boundaries

- Repo-local analysis only.
- No shell startup file parsing.
- No env loading or injection.
- No secret manager integration.
