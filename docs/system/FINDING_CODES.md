# Finding Codes

Stable finding-code reference for `envdiff doctor`.

## Deterministic Findings

| Code | Severity | Meaning |
| --- | --- | --- |
| `ENV001` | `error` | Required variable is referenced but missing from the associated `.env` |
| `ENV002` | `warning` | Optional variable is referenced but missing from the associated `.env` |
| `ENV003` | `info` | Committed `.env` variable is defined but not referenced |
| `ENV004` | `warning` | Referenced variable is missing from the nearest `.env.example` |
| `ENV005` | `info` | Template-only `.env.example` entry appears stale |
| `ENV006` | `warning` | Variable is defined more than once in the same dotenv file |

## Heuristic Findings

| Code | Severity | Meaning |
| --- | --- | --- |
| `ENV007` | `warning` | Missing variable has a likely nearby alias candidate |
| `ENV008` | `warning` | Committed `.env` value looks like a real secret |
| `ENV009` | `warning` | Committed `.env` value looks like a placeholder |

## Notes

- Codes are intended to remain stable for suppression, baselines, and CI use.
- `severity` is part of the public contract.
- Heuristic findings include `reason`, `confidence`, and `source_kind`.
- `suppression_key` is the stable identifier for ignore files and baseline snapshots.

## Code Change Procedure

Finding codes are a public contract. Baselines, ignore files, and CI policies can
key on them.

- Do not reuse a code for a different meaning.
- Prefer adding a new code over changing the meaning of an existing code.
- Changing severity is a compatibility change; update docs and tests in the same
  commit.
- If a suppression key format changes, document the migration path because existing
  baselines and ignore files may stop suppressing findings.

When adding or changing a finding:

1. Update the analyzer that emits it.
2. Update this file and `docs/system/FEATURES.md`.
3. Add focused analyzer tests and a CLI smoke test when user-visible output changes.
4. Keep finding ordering deterministic through `src/utils/ordering.py`.
5. Run the full local gate: `uv run ruff check .`, `uv run ruff format --check .`,
   and `uv run pytest -q`.
