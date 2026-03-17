# envdiff

`envdiff` is a local-first CLI that helps you understand and validate your environment variable contract across dotenv files, source code, and deployment-oriented config.

Instead of just diffing two `.env` files, it answers higher-value questions:

- Which variables are required by the codebase?
- Which are missing, stale, or undocumented?
- Where do names drift across repos and services?
- Which committed values look like real secrets?
- How different are dev, staging, and prod contracts?

The tool is designed for engineers, platform teams, and AI agents that need deterministic configuration analysis.

---

## Why this exists

Environment configuration tends to decay quietly:

- `.env.example` goes stale
- old variables stick around forever
- new required variables are added in code but never documented
- staging/prod/local use different names for the same thing
- dangerous placeholders or real secrets end up committed
- onboarding turns into configuration archaeology

`envdiff` treats env vars as an interface contract and makes drift visible.

---

## Core principles

- **local-first** — useful without network access
- **deterministic** — same inputs produce the same outputs
- **agent-friendly** — stable JSON output for automation
- **practical** — optimized for real repo hygiene, not novelty
- **incremental** — starts with core parsers and expands over time

---

## What it does

### Compare env files
```bash
envdiff compare .env .env.example
```

Detects:
- missing keys
- extra keys
- duplicate definitions
- value-class differences

### Scan a repo
```bash
envdiff scan .
```

Infers env usage from:
- `.env` / `.env.example`
- Python env access
- Docker Compose interpolation

Reports:
- referenced vars
- defined vars
- required vs optional
- source locations
- undocumented variables

### Validate contract health
```bash
envdiff doctor .
```

Flags:
- missing required variables
- missing optional variables
- unused variables
- alias candidates / naming drift
- secret-like committed values
- placeholders and suspicious defaults

---

## Example

Given this `.env.example`:

```env
DATABASE_URL=
REDIS_URL=
DEBUG=false
OPENAI_API_KEY=
```

This local `.env`:

```env
DATABASE_URL=postgres://localhost/db
DEBUG=true
OPENAI_KEY=abc123
OLD_CACHE_URL=redis://localhost:6379
```

And Python code that references:

```python
db = os.environ["DATABASE_URL"]
redis = os.getenv("REDIS_URL")
openai = os.environ["OPENAI_API_KEY"]
```

A doctor run could report:

- missing required: `OPENAI_API_KEY`
- missing optional: `REDIS_URL`
- possibly misnamed: `OPENAI_KEY` → `OPENAI_API_KEY`
- unused: `OLD_CACHE_URL`
- local skew: `DEBUG=true` differs from example default `false`

---

## Tech stack

- **Python 3.11+**
- **uv** for project and dependency management
- **Typer** for CLI ergonomics
- **Rich** for terminal rendering
- **Pydantic** for internal data models
- Python AST for source inference
- heuristic analyzers for aliases and secret detection

Core functionality should rely on the standard library where practical.

---

## Repository architecture

```text
envdiff/
├── README.md
├── pyproject.toml
├── uv.lock
│
├── envdiff/
│   ├── cli.py
│   ├── models.py
│   │
│   ├── parsers/
│   │   ├── dotenv.py
│   │   ├── python_ast.py
│   │   └── compose.py
│   │
│   ├── analyzers/
│   │   ├── compare.py
│   │   ├── scan.py
│   │   ├── doctor.py
│   │   ├── aliases.py
│   │   └── secrets.py
│   │
│   ├── render/
│   │   ├── human.py
│   │   └── json.py
│   │
│   └── utils/
│       ├── entropy.py
│       ├── normalize.py
│       ├── paths.py
│       └── severity.py
│
└── tests/
    ├── fixtures/
    ├── test_dotenv.py
    ├── test_python_ast.py
    ├── test_compose.py
    ├── test_compare.py
    └── test_doctor.py
```

---

## Initial scope

### Supported in v1
- dotenv parsing
- Python env usage inference
- Docker Compose interpolation
- compare
- scan
- doctor
- deterministic JSON output
- alias and secret heuristics

### Deferred
- GitHub Actions
- shell scripts
- Pydantic/BaseSettings inference
- Kubernetes/Helm/Terraform
- secret managers
- automatic rewrites

---

## Output expectations

### Human mode
Readable terminal summaries with grouped findings:
- Summary
- Missing
- Unused
- Alias candidates
- Secret hygiene
- Suggestions

### JSON mode
All major commands should support:
```bash
envdiff doctor . --json
```

JSON is intended for:
- CI/CD
- editor tooling
- autonomous terminal agents
- scripted reporting

---

## Exit codes

Suggested behavior:

- `0` — success, no findings at or above threshold
- `1` — execution/parsing error
- `2` — findings at or above `--fail-on`

Example:
```bash
envdiff doctor . --fail-on warning
```

---

## Long-term direction

Over time, `envdiff` can evolve from a repo hygiene utility into a broader configuration contract tool with support for:

- multi-environment matrix analysis
- `.env.example` generation
- monorepo/service grouping
- CI policy enforcement
- baseline drift snapshots
- broader parser ecosystems

The core constraint should remain the same: deterministic, practical, agent-usable analysis.

---

## Status

Planned. Initial implementation should prioritize:
1. compare
2. scan
3. doctor

Everything else is downstream of a solid contract model.
