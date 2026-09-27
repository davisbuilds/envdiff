# Git History and Branch Hygiene

Last updated: September 27, 2026

## Repository Merge Settings

Configured on GitHub repository `davisbuilds/envdiff`:

- `allow_squash_merge`: `false`
- `allow_merge_commit`: `true`
- `allow_rebase_merge`: `true`
- `delete_branch_on_merge`: `true`
- `merge_commit_title`: `PR_TITLE`
- `merge_commit_message`: `PR_BODY`

Result:

- PR branches retain their full commit history when merged.
- `main` receives either a merge commit (preserving the PR boundary) or rebased commits (linear history), depending on which strategy the merger picks for that PR.
- Squash merging is disabled — full per-commit history is preserved.
- Merged remote branches are auto-deleted.

## Merge Strategy

Merge commits and rebase merges are both allowed; squash merges are disabled.

- **Default — merge commit.** Preserves the PR as a discoverable boundary in `main`'s history. Best when the PR contains multiple meaningful commits worth keeping addressable individually (e.g. the layered Go port).
- **Rebase merge.** Use when the PR's commits are clean and the linear history reads better without an extra merge node. Avoid if the PR's commits are noisy (WIP, fixups) — clean them up locally first.
- **Authoring expectation.** Because squash is gone, individual PR commits land in `main`. Keep PR commit messages tidy: meaningful subjects, no WIP markers, no fixup chains. Squash or reword locally before opening the PR if needed.

## Conventional Commits and Releases

Merge and rebase retain individual commits, so every non-merge PR commit must
use `type(scope)?: description` (optional scope and optional `!` before the
colon). Use `feat`, `fix`, `perf`, `docs`, `test`, `chore`, `build`, `ci`, `style`,
`refactor`, or `revert`. PR titles use the same syntax because merge commits use
`PR_TITLE`. Actual merge commits inside a branch are excluded from the subject
check. Reword WIP/fixup commits before opening the PR.

`feat` requests a minor release; `fix`/`perf` request a patch. Mark breaking
changes with `!` or a `BREAKING CHANGE:` footer, even before 1.0, and explain the
migration. Before 1.0, breaking changes bump minor. Syntax validation does not
verify compatibility classification. Release PRs and published notes still need
review; [release operations](../system/RELEASES.md) own the release contract.

CI checks PR titles on edits and all retained non-merge subjects. Main-push CI
also checks non-merge commits in the pushed `before..head` range, including direct
pushes, and all unreleased commits before it can authorize release work. The
durable boundary is the actual `vX.Y.Z` tag matching the release manifest, or the
configured bootstrap SHA when no matching tag exists. A release PR's not-yet-
tagged manifest version uses that conservative fallback. A later valid push
cannot hide an earlier invalid unreleased main commit. Missing, zero, or unavailable
revision SHAs, identical ranges, and non-forward pushes fail closed;
pre-bootstrap history is not traversed. The ancestry requirement is push-only:
a PR may diverge from an advanced base. Locally run:

```bash
PR_TITLE='feat: describe the change' python3 scripts/check_release_commits.py \
  "$(git rev-parse origin/main)" "$(git rev-parse HEAD)" --check-pr-title
```

Old main history predating the PR is outside this check.

## Branch Protection

Enabled on `main` (this is a public repository, so protection APIs are available):

- `required_conversation_resolution`: `true` — all PR review threads must be resolved before merge.
- `enforce_admins`: `false` — admins may still push directly to `main` for trivial maintenance.
- No required reviews or status checks are enforced as branch rules yet; CI gates below are enforced by convention.

## CI Gates

Workflow: `.github/workflows/ci.yml`

Quality gates expected green before merge (matches the project pre-push checklist):

- `go vet ./...`
- `go test ./...`
- `golangci-lint run ./...`
- release validator controls: `python3 -m unittest discover -s scripts -p 'test_release_commits.py'`
- PR title and retained commit categories; main-push commit categories
- goldens current: `ENVDIFF_UPDATE_GOLDENS=1 go test ./... && git diff --exit-code tests/golden`

## Recommended Ongoing Hygiene

1. Create short-lived feature branches from `main` (`feat/*`, `fix/*`, `docs/*`, `chore/*`).
2. Open PRs early; keep them focused on one intent.
3. Tidy your PR commit history *before* merging — reword/squash locally so what lands on `main` reads cleanly.
4. Pick **Create a merge commit** by default; pick **Rebase and merge** when linear history is materially better.
5. Resolve all review conversations before merging (enforced by branch protection).
6. Periodically prune local branches:

```bash
git fetch --prune
git branch --merged main | grep -v ' main$' | xargs -n 1 git branch -d
```
