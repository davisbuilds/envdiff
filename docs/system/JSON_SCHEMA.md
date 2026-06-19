# JSON Schema

Stable machine-readable contract for `envdiff --json`.

## Envelope

All JSON commands emit the same top-level envelope:

```json
{
  "meta": {
    "command": "doctor",
    "schema_version": "1"
  },
  "inputs": {},
  "summary": {
    "error": 0,
    "warning": 0,
    "info": 0
  },
  "findings": [],
  "data": {}
}
```

## Top-Level Fields

| Field | Type | Meaning |
| --- | --- | --- |
| `meta.command` | string | Active command name |
| `meta.schema_version` | string | Machine contract version |
| `inputs` | object | Normalized command inputs |
| `summary` | object | Severity counts |
| `findings` | array | Structured findings, empty for non-finding commands |
| `data` | object | Command-specific payload |

## Finding Object

`doctor` emits `Finding` objects under `findings[]`:

| Field | Type | Meaning |
| --- | --- | --- |
| `code` | string | Stable finding code like `ENV001` |
| `severity` | string | `error`, `warning`, or `info` |
| `title` | string | Short classification label |
| `details` | string | Human-readable explanation |
| `variable_name` | string or null | Primary variable involved |
| `locations` | array | Source locations |
| `related_variables` | array | Heuristic-related names |
| `suggested_fix` | string or null | Optional remediation hint |
| `confidence` | string or null | Heuristic confidence label |
| `source_kind` | string | `deterministic` or `heuristic` |
| `reason` | string or null | Explainable rationale |
| `suppression_key` | string or null | Stable suppression token |

## Command Data Shapes

### `compare`

`data` includes:
- `left_path`
- `right_path`
- `missing_in_left`
- `missing_in_right`
- `duplicates_in_left`
- `duplicates_in_right`
- `differing_values`
- `warnings`

### `scan`

`data` is the serialized `RepoScanResult`:
- `root_path`
- `definitions`
- `usages`
- `contracts`
- `resolutions`
- `warnings`

### `matrix`

`data` includes:
- `paths`
- `show_all`
- `file_count`
- `variable_count`
- `inconsistent_variable_count`
- `variables`
- `warnings`

Each `variables[]` entry includes:
- `name`
- `status`
- `present_in`
- `missing_in`
- `duplicates_in`
- `value_kinds`
- `files`

### `generate`

`data` includes:
- `root_path`
- `annotate`
- `variable_count`
- `variables`
- `generated_text`
- `output_path`
- `check`

`check` is either `null` or:

```json
{
  "target_path": "/abs/or/relative/path",
  "exists": true,
  "matches": false
}
```

### `doctor`

`data` includes:
- `scan`
- `filtering`
- `suppressed_findings`

`filtering` includes:
- `baseline_entries`
- `suppressed_count`
- `baseline_written`

## Stability Notes

- Ordering is deterministic.
- New fields may be added in a future schema version.
- Existing fields in schema version `1` should be treated as stable.
- Consumers should key on `meta.command` and `meta.schema_version`.
- **Encoding:** output is UTF-8 with two-space indentation and lexicographically
  sorted keys. Non-ASCII values are emitted as raw UTF-8 (not `\uXXXX` escapes),
  and `<`, `>`, `&` appear literally rather than HTML-escaped.

## Contract Change Procedure

The JSON envelope is a public machine contract. Treat changes as compatibility work,
not internal refactors.

Schema version can stay the same when:

- Adding optional fields that older consumers can ignore.
- Adding a new command-specific `data` field without removing or renaming existing
  fields.
- Adding new warning strings while preserving existing structured fields.

The **Go implementation is the contract source**; the Python package is a
transitional parity oracle. Goldens under `tests/golden/json/` are generated
from the Go binary (`uv run python scripts/update_go_golden.py`).

Bump `SchemaVersion` in `internal/version` (and the mirrored `SCHEMA_VERSION` in
`src/envdiff/models.py` while the oracle exists) when:

- Renaming or removing an existing field.
- Changing a field's type or nullability.
- Changing top-level envelope shape.
- Changing deterministic ordering semantics.
- Changing `summary`, `findings`, or `data` in a way an existing JSON consumer could
  misinterpret.

For any schema-affecting change:

1. Update the Go models in `internal/model` (and `internal/version` for a version
   bump).
2. Update JSON rendering in `internal/render` if needed.
3. Regenerate goldens from Go: `uv run python scripts/update_go_golden.py`.
4. Update this file and affected command docs.
5. While the Python oracle exists, mirror the change in `src/envdiff/models.py`
   and `src/envdiff/render/json.py` so the parity gate stays green.
6. Run the full gate: `go test ./...`, `uv run pytest -q`,
   `uv run python scripts/check_go_parity.py`, `uv run ruff check .`, and
   `uv run ruff format --check .`.

Do not silently change field names to make human output cleaner. Human rendering and
JSON rendering have different compatibility requirements.
