# Features

Current product-surface reference for `envdiff`.

## Supported Inputs

- `.env`
- `.env.example`
- Python:
  - `os.environ["X"]`
  - `os.getenv("X")`
  - `os.getenv("X", "default")`
- Docker Compose:
  - `${VAR}`
  - `${VAR:-default}`

## Commands

| Command | Purpose | Current Notes |
| --- | --- | --- |
| `compare` | Compare two dotenv files | Reports missing keys, duplicates, and value-kind differences |
| `scan` | Analyze a repo's env contract surface | Aggregates definitions, usages, and repo-local resolution |
| `doctor` | Validate the inferred contract | Emits structured findings, supports `--fail-on`, and can baseline or suppress findings |

## Finding Surface

| Code | Meaning | Type |
| --- | --- | --- |
| `ENV001` | Missing required variable | Deterministic |
| `ENV002` | Missing optional variable | Deterministic |
| `ENV003` | Unused committed variable | Deterministic |
| `ENV004` | Referenced variable missing from nearest `.env.example` | Deterministic |
| `ENV005` | Stale template-only `.env.example` entry | Deterministic |
| `ENV006` | Duplicate definition | Deterministic |
| `ENV007` | Possible alias candidate | Heuristic |
| `ENV008` | Secret-like committed value | Heuristic |
| `ENV009` | Placeholder-like committed value | Heuristic |

## Current Heuristics

### Alias / Drift Detection

- conservative, low-confidence only
- explainable reason strings
- focused on nearby defined names for missing variables
- supports common forms such as `DB_URL` vs `DATABASE_URL`, `PGHOST` vs `POSTGRES_HOST`, and `OPENAI_KEY` vs `OPENAI_API_KEY`

### Secret / Placeholder Hygiene

- flags secret-like values only in committed `.env` files
- flags non-empty placeholder values like `changeme`
- keeps `.env.example` placeholders out of the warning path

## Output Modes

### Human Output

- concise command summaries
- grouped doctor findings by severity
- summary counts and suppression counts when applicable

### JSON Output

- schema-versioned envelope
- explicit `meta`, `inputs`, `summary`, `findings`, and `data` fields
- deterministic ordering for contracts and findings
- suppression metadata and suppressed finding lists for `doctor`

## Deferred Features

- GitHub Actions parsing
- shell script or `.envrc` parsing
- matrix mode
- `.env.example` generation
- secret manager integration
