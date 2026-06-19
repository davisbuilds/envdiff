# JSON Goldens

These JSON files are generated from the Go `envdiff` binary, the source of truth
for the JSON contract, and are asserted against by the Go test suite. Values are
emitted as raw UTF-8 (non-ASCII is not `\uXXXX`-escaped).

Regenerate after intentional Go behavior changes:

```bash
uv run python scripts/update_go_golden.py
```

Check committed files without rewriting:

```bash
uv run python scripts/update_go_golden.py --check
```

The generator normalizes machine-local absolute paths before writing files:

- `<REPO_ROOT>` replaces the local checkout path.
- `<TMPDIR>` replaces temporary directories used for baseline suppression cases.

Tests compare decoded JSON values after applying the same normalization, not raw
rendered JSON strings — so byte-level encoding choices (raw UTF-8, key order) are
invisible to them. `TestScanRepositoryEmitsRawUTF8` pins the raw-UTF-8 contract
at the byte level via the `unicode_repo` fixture.
