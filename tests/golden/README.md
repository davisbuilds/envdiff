# Go Port Goldens

These JSON files are generated from the current Python `./envdiff` CLI and act as
the oracle for the side-by-side Go port.

Regenerate after intentional Python behavior changes:

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

Tests should compare decoded JSON values after applying the same normalization,
not raw rendered JSON strings.
