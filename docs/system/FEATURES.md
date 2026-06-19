# Features

Current product-surface reference for `envdiff`.

## Supported Inputs

- `.env`
- `.env.example`
- Python:
  - `os.environ["X"]`
  - `os.getenv("X")`
  - `os.getenv("X", "default")`
- JavaScript/TypeScript (`.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs`):
  - `process.env.X` and `process.env["X"]` (optional — JS returns `undefined`)
  - `process.env.X || "default"` / `process.env.X ?? "default"`
    (optional with default)
- Shell (`.sh`, `.bash`):
  - `$VAR` / `${VAR}` (required), `${VAR:-default}` / `${VAR:=default}`
    (optional with default), `${VAR:?msg}` (required)
  - `export VAR=…` / `VAR=…` are treated as local definitions and suppress
    same-file "missing" findings only (they do not satisfy `.env` contracts)
- Dockerfile (`Dockerfile`, `Dockerfile.*`, `*.Dockerfile`):
  - `$VAR` / `${VAR}` interpolation (same requiredness rules as shell)
  - `ARG`/`ENV` are local definitions (same-file suppression, separate scope)
- Docker Compose:
  - `${VAR}`
  - `${VAR:-default}`
- GitHub Actions:
  - `${{ secrets.NAME }}`
  - `${{ vars.NAME }}`
  - `${{ vars.NAME || 'default' }}`

## Commands

| Command | Purpose | Current Notes |
| --- | --- | --- |
| `compare` | Compare two dotenv files | Reports missing keys, duplicates, and value-kind differences |
| `generate` | Produce a `.env.example` candidate | Prints generated dotenv to stdout, writes via `--output`, or checks drift with `--check` |
| `matrix` | Compare multiple dotenv files | Reports presence, kind mismatches, and duplicate signals across two or more files |
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
- generated dotenv content to stdout by default, with summary output on explicit writes
- `generate --check` reports whether the committed template matches inferred output
- mismatch-focused matrix summaries across multiple dotenv files
- grouped doctor findings by severity
- summary counts and suppression counts when applicable

### JSON Output

- schema-versioned envelope
- explicit `meta`, `inputs`, `summary`, `findings`, and `data` fields
- deterministic ordering for contracts and findings
- generated dotenv text and output-path metadata for `generate`
- generate drift-check metadata for CI use
- suppression metadata and suppressed finding lists for `doctor`

## Deferred Features

- shell script or `.envrc` parsing
- secret manager integration
