# AGENTS.md

`envdiff` is a deterministic Go CLI that scans repos for environment variable
contracts, compares `.env` files, and flags mismatches. Local-first; no network
in core analysis. `./envdiff --help` lists commands (the launcher builds and
caches `bin/envdiff`).

## Documentation Map

- `CONTRIBUTING.md` — contribution scope and delivery.
- `docs/system/ARCHITECTURE.md` — flow, layers, directory map.
- `docs/system/FEATURES.md` — inputs, commands, heuristics, output modes, deferred features.
- `docs/system/OPERATIONS.md` — setup, runnable fixture repos, exit codes, constraints.
- `docs/system/JSON_SCHEMA.md`, `docs/system/FINDING_CODES.md` — the public JSON and finding-code contracts.
- `docs/system/RELEASES.md` — SemVer, release automation, recovery.
- `docs/project/SPEC.md` (scope), `ROADMAP.md`, `BACKLOG.md`.
- `docs/project/GIT_HISTORY_POLICY.md` — merge settings, Conventional Commits, branch protection.

## Project Boundaries

- Keep it a deterministic, repo-local env contract analyzer: not an env loader,
  injector, or secret-distribution tool, and user-shell startup files are out of
  scope. Prefer conservative heuristics over aggressive recall.
- **The JSON envelope, field names, finding codes, CLI arguments, output streams,
  and exit codes are public contracts.** Keep finding and contract ordering
  deterministic. A change to any of them updates `docs/system/` in the same change
  and is marked breaking (`!` or a `BREAKING CHANGE:` footer) when incompatible.

## Testing

- **Local gate** (matches the required `Build & test`, `Lint & format`, `check`,
  and `Release commit categories` checks, plus `security.yml`):

  ```bash
  go mod tidy -diff && go mod verify && go vet ./... && go test -race ./... \
    && golangci-lint run ./... \
    && ENVDIFF_UPDATE_GOLDENS=1 go test ./... >/dev/null && git diff --exit-code tests/golden \
    && go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./... \
    && python3 scripts/check_publication_hygiene.py \
    && python3 scripts/check_comments.py \
    && python3 -m unittest discover -s scripts -p 'test_release_commits.py'
  ```

  CI also validates the PR title and every retained commit subject; run the
  `check_release_commits.py` command in `docs/project/GIT_HISTORY_POLICY.md`.
  For workflow changes, also run `uvx zizmor@1.30.0 --offline .github/workflows/`.
- Use real fixture repos under `tests/fixtures/`, not mocks. Parser work gets
  focused tests (`internal/parsers`, `internal/dotenv`) plus a repo-scan
  integration test (`internal/analyzers/scan_test.go`); CLI changes update
  `internal/cli/cli_test.go`.
- **JSON goldens** (`tests/golden/json/`) are the rendered-output contract. Each
  golden has exactly one writer, an analyzer-level test via
  `testutil.AssertGoldenJSON`; CLI and render tests only consume them.
  `TestScanRepositoryEmitsRawUTF8` pins raw UTF-8 at the byte level.
- `golangci-lint` (`.golangci.yml`, including `gofumpt`) owns style and dead code;
  fix what it flags.

## Working Agreement

- **Clarify before building** when a request is vague, contradictory, or skips a
  key decision; clear, well-scoped requests need no interview.
- **Fix small issues now; backlog the rest.** Fix simple, quick, or blocking
  issues in this repo as part of the current change and call them out. Record in
  `docs/project/BACKLOG.md` only what can't be finished now or needs the
  maintainer's decision; don't open a branch or PR just for a backlog entry.
- **Keep docs current** when a change makes a documented contract, boundary,
  procedure, or direction stale.
- **Plan and hand off work.** Reuse clear task context. Save planning docs when
  decisions, acceptance, or coordination need persistence, and leave a brief
  handoff only when needed to resume unfinished work. Commit shared docs and keep
  status current. `docs/archive/` is gitignored: before archiving a completed doc
  there, preserve lasting decisions and open work in maintained docs and repair
  links.
- **TDD:** red/green for features, major refactors and large changes; the red
  test must fail on the behavior, not on a missing symbol. For smaller edits, run
  the relevant existing checks.
- **Comments.** Don't add code comments. The one exception is a decision record:
  start it with `WHY:` and state a non-obvious reason the code can't express.
  Never justify a shortcut in a comment; fix it or record it in
  `docs/project/BACKLOG.md`. `python3 scripts/check_comments.py` enforces a
  ratchet: after deleting comments run it with `--update`, and never raise a count
  in `.comment-baseline.json`.
- **Commit and push.** Commit coherent chunks, staging by explicit path. Push
  your task branch and open a PR once the work is complete and the local gate
  passes; don't push directly to `main`.
- **Merge** only under the standing merge authority in the maintainer's workspace
  instructions, or when the maintainer asks. Without those instructions in
  context, leave the PR for the maintainer.
- **History and releases.** Squash is disabled: merge commit by default, rebase
  only when the branch's commits are already clean. Every retained commit
  subject and the PR title use Conventional Commits, because Release Please
  reads them. Merging to `main` updates the Release Please PR; merging that PR
  tags a GitHub Release, so leave it for the maintainer.
- **Re-ground after compaction** in source, this file, the owning docs, and
  recent commits. A summary and your own earlier proposals are not user approval.
