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
