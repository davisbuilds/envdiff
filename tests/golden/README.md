# JSON Goldens

These JSON files are generated from the Go `envdiff` binary, the source of truth
for the JSON contract, and are asserted against by the Go test suite. Values are
emitted as raw UTF-8 (non-ASCII is not `\uXXXX`-escaped).

Regenerate after intentional Go behavior changes by running the test suite with
the update flag set:

```bash
ENVDIFF_UPDATE_GOLDENS=1 go test ./...
```

Each golden has exactly one writer (an analyzer-level test that renders the
command output); the CLI and render tests are pure consumers that validate the
committed files. A normal `go test ./...` checks the goldens without rewriting.

Goldens normalize the machine-local checkout path to `<REPO_ROOT>` before
writing. Tests compare decoded JSON values after applying the same
normalization, not raw rendered JSON strings — so byte-level encoding choices
(raw UTF-8, key order) are invisible to them. `TestScanRepositoryEmitsRawUTF8`
pins the raw-UTF-8 contract at the byte level via the `unicode_repo` fixture.
