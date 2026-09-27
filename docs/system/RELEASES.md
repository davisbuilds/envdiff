# Releases

The release unit is the Go CLI source, identified by `vX.Y.Z` tags and GitHub
Releases. Consumers clone a tag and build with `go build -o bin/envdiff
./cmd/envdiff` or use the local `./envdiff` launcher. This automation does not
publish a registry package or attach prebuilt binaries.

## Version sources and compatibility

`internal/version/version.go` owns the application `Version`; `envdiff --version`
prints `envdiff X.Y.Z` to stdout and exits 0. It accepts no other arguments and
has no JSON mode. Builds between releases report the checked-in version, so use
the Git commit when identifying development changes.

`SchemaVersion = "1"` separately owns the JSON envelope contract. Application
releases do not change it; JSON goldens continue to enforce the existing schema.
CLI arguments, stdout/stderr, exit codes, finding semantics, and JSON fields are
compatibility surfaces. Mark intentional incompatibility with `!` or a
`BREAKING CHANGE:` footer and describe migration steps.

Before 1.0, features and breaking changes bump the minor version; fixes and
performance improvements bump the patch. A 1.0 release requires an explicit
compatibility decision. Documentation, tests, CI, and maintenance alone do not
normally request a consumer release.

## Bootstrap

The September 27, 2026 inventory found no existing tags or GitHub Releases.
`0.0.0` in the application constant and `.release-please-manifest.json` is an
unreleased sentinel, not a past release. The first release is configured as
`0.1.0`. The initial commit comparison starts after actual main commit
`4f2b8a852bc0451a4a569cd5817b0ef56272f3c4`; older work is not retroactively
released or summarized. Review the first PR body and changelog against that
commit boundary, including the newly added application version flag.

## Release workflow

Release Please proposes a PR updating the application constant, manifest, and
`CHANGELOG.md`. Review its version, changelog, PR body, and compatibility intent;
merge only after normal checks and review. After the release PR merges and its
main-push CI passes, the workflow creates its tag and GitHub Release. It is the
single release/tag writer; do not create competing manual tags.

`.github/workflows/release-please.yml` accepts only successful same-repository
main-push runs of `CI`. It checks current main against the tested SHA before
acquiring a repository-scoped App token. That check is a point-in-time preflight,
not a lock against subsequent pushes. Release writes serialize without cancelling
an active writer. The privileged workflow does not check out repository code or
load event artifacts/caches.

Setup requires the release App installed on this repository, Actions variable
`RELEASE_APP_CLIENT_ID`, and secret `RELEASE_APP_PRIVATE_KEY`. It requests only
Contents, Issues, and Pull requests write. Keep the private key out of source and
logs. Inspect the hosted release run, generated PR changes, and automatically
started PR CI to establish activation; stored credentials alone do not prove it.

## Checks and recovery

Run the existing Go vet, test, lint, and golden checks, plus
`python3 -m unittest discover -s scripts -p 'test_release_commits.py'` and
`uvx zizmor@1.30.0 --offline .github/workflows/` for workflow changes. CI validates
retained Conventional Commit subjects and the PR title, plus main-push commits
(including direct pushes) before release authority; push ranges must be nonempty
and advance from an ancestor revision. Push CI also validates all unreleased
commits after the actual tag matching the manifest, or the configured bootstrap
SHA if the tag does not exist. Later pushes cannot forget an earlier failed
classification. See the Git history
policy. Syntax checks cannot decide whether a change breaks compatibility.

If automation fails, inspect the run before retrying. When it may have created a
PR, tag, or release, inspect that remote state first to avoid a duplicate writer.
Fix workflow/configuration problems in a normal reviewed PR. Correct proposed
notes and compatibility mistakes before merging the release PR. Do not rewrite
published tags; use a corrective release for published mistakes.

If an unclassified commit is already on main, stop release work and review all
unreleased changes and their compatibility intent. Preserve published history.
Any rewrite of unpublished main history requires an explicit owner decision.
Otherwise keep automation blocked pending an owner-approved recovery release
whose version and notes cover the complete unreleased range. Pause this writer
before any separately authorized manual recovery publication; do not fabricate a
historical tag or silently skip the failed commit. Resume only when the application
version and manifest match that actual reviewed release tag. Already published
commits are outside the gate when the manifest matches their real release tag.
