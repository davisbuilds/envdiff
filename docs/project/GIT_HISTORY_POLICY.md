# Git History and Branch Hygiene

Last updated: June 18, 2026

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

## Branch Protection

Enabled on `main` (this is a public repository, so protection APIs are available):

- `required_conversation_resolution`: `true` — all PR review threads must be resolved before merge.
- `enforce_admins`: `false` — admins may still push directly to `main` for trivial maintenance.
- No required reviews or status checks are enforced as branch rules yet; CI gates below are enforced by convention.

## CI Gates

Workflow: `.github/workflows/ci.yml`

Quality gates expected green before merge (matches the project pre-push checklist):

- `uv run ruff check .`
- `uv run ruff format --check .`
- `uv run pytest -q`
- `go test ./...`
- `uv run python scripts/check_go_parity.py`

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
