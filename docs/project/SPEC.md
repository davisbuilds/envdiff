# envdiff Specification

## Overview

`envdiff` is a local-first CLI for understanding, validating, and comparing environment configuration contracts across source code, dotenv files, and deployment-oriented config. It is designed to detect configuration drift, missing variables, dead variables, naming inconsistencies, unsafe defaults, and secret hygiene issues.

The core concept is that environment variables are not just strings in files; they are an interface contract between application code, infrastructure, CI/CD, and operators.

`envdiff` is a contract analyzer. It is not an environment loader, secret distribution system, runtime injector, or shell session manager. The product boundary should stay centered on static, repo-local analysis and deterministic diagnostics.

---

## Goals

- Detect missing, extra, unused, and inconsistent environment variables.
- Infer expected environment variables from application code and config.
- Compare environment definitions across multiple files and environments.
- Surface likely naming drift and alias candidates.
- Flag suspicious secret exposure and unsafe placeholders.
- Produce deterministic human-readable and JSON outputs for terminal users and AI agents.
- Remain useful fully offline.

---

## Non-Goals (v1)

- Secret retrieval from cloud secret managers.
- Secret encryption, distribution, and key-management workflows.
- Runtime env loading or process injection.
- Runtime environment introspection from deployed clusters.
- Full shell semantic evaluation.
- User shell startup file analysis such as `~/.zshrc`, `~/.zprofile`, or `~/.bashrc`.
- Automatic code refactors.
- Kubernetes, Terraform, Helm, and cloud platform support in v1.
- Full multi-language source parsing beyond the initial supported set.

---

## Target Users

- Individual developers managing local and project env files.
- Small teams dealing with `.env.example`, CI, and Docker Compose drift.
- Platform / DevOps engineers validating repo configuration contracts.
- AI coding agents needing deterministic env validation.

---

## Core User Jobs

1. Compare two or more env files and understand what differs.
2. Scan a repo to infer which variables are expected by the codebase.
3. Detect stale or undocumented variables.
4. Find likely naming mismatches such as `DB_URL` vs `DATABASE_URL`.
5. Validate whether an env file is safe and complete enough for a target environment.
6. Generate a clean `.env.example` from observed usage.

---

## v1 Command Surface

### `envdiff compare`
Compare two dotenv-style files.

Example:
```bash
envdiff compare .env .env.example
```

Expected output:
- Missing keys in left/right file
- Extra keys in left/right file
- Duplicate definitions
- Normalized value-class differences
- Optional JSON output

---

### `envdiff scan`
Scan a repository for env variable usage.

Example:
```bash
envdiff scan .
```

Supported v1 sources:
- `.env`
- `.env.example`
- Python code:
  - `os.environ["X"]`
  - `os.getenv("X")`
  - `os.getenv("X", "default")`
- Docker Compose:
  - `${VAR}`
  - `${VAR:-default}`

Expected output:
- Referenced variables
- Defined variables
- Missing definitions
- Undocumented variables
- Required vs optional inference
- Source locations
- Repo-local contract resolution details where relevant

---

### `envdiff doctor`
Run contract validation and linting.

Example:
```bash
envdiff doctor .
```

Checks:
- Missing required vars
- Missing optional vars
- Unused vars
- Possibly misnamed vars
- Secret-like committed values
- Unsafe placeholders
- Environment skew vs example definitions

Supports:
```bash
envdiff doctor . --json
envdiff doctor . --fail-on warning
```

---

### Optional v1.1 commands

### `envdiff matrix`
Compare multiple environments.

Example:
```bash
envdiff matrix envs/*.env
```

### `envdiff generate`
Generate `.env.example` from inferred contract.

Example:
```bash
envdiff generate . --annotate
```

These should not block v1.

---

## Functional Requirements

### 1. Dotenv Parsing
Must parse:
- `KEY=value`
- quoted values
- blank values
- comments
- duplicate keys
- commented-out definitions as non-active lines

Must preserve:
- key name
- value
- line number
- file path
- duplicate count
- parse warnings

---

### 2. Python Usage Inference
Must detect:
- `os.environ["X"]` as required
- `os.getenv("X")` as optional/unknown
- `os.getenv("X", default)` as optional with default

Should capture:
- variable name
- usage type
- file path
- line number
- inferred requiredness

Implementation may use Python AST rather than regex.

---

### 3. Docker Compose Usage Inference
Must detect:
- `${VAR}`
- `${VAR:-default}`

Should infer:
- required if no default
- optional if default present

Must record source file and location where feasible.

---

### 4. Comparison Engine
Must detect:
- missing keys
- extra keys
- duplicate keys
- normalized value differences
- conflicting defaults

Normalized value classes should include:
- boolean
- integer
- float
- URL
- secret-like opaque string
- placeholder
- generic string

The tool should compare value class before raw value when producing summaries.

---

### 5. Contract Analyzer
Given inferred usage and definitions, must classify each variable into one or more categories:
- referenced and defined
- referenced but undefined
- defined but unreferenced
- required
- optional
- optional with default
- secret-like
- placeholder-like
- alias candidate

---

### 6. Repository Resolution Model
The tool must define deterministic repo-local resolution rules for associating definitions, examples, and usages.

v1 should support:
- deterministic directory traversal
- ignore-path support
- nearest relevant `.env` / `.env.example` association where repo layout requires disambiguation
- explicit fallback behavior when multiple candidate env files exist
- recorded provenance for why a given definition/example file was associated with a usage set

The resolution model must be documented because it affects correctness in monorepos and multi-service repos.

---

### 7. Alias / Drift Detection
Must detect likely name drift via heuristics:
- edit distance
- token similarity
- suffix/prefix similarity
- acronym vs expanded form

Examples:
- `DB_URL` ↔ `DATABASE_URL`
- `OPENAI_KEY` ↔ `OPENAI_API_KEY`
- `PGHOST` ↔ `POSTGRES_HOST`

Alias detection must remain heuristic and low-confidence by default.
Alias findings must include an explainable reason, such as token overlap or acronym expansion.

---

### 8. Secret Hygiene Detection
Must flag values that appear to be:
- API tokens
- private keys
- long opaque secrets
- JWT-like blobs

Must distinguish likely placeholders such as:
- `changeme`
- `your_key_here`
- `example`
- empty string

v1 should use heuristics only. No network calls.

---

### 9. Heuristic Safety and Explainability
Heuristic-based findings must favor low false-positive rates over aggressive recall in v1.

The system must provide:
- fixed thresholds
- deterministic ordering
- explainable finding reasons
- stable finding codes suitable for future suppression or baseline workflows
- confidence levels where heuristics are involved

Placeholder-like values should not be misclassified as secrets by default.

---

### 10. JSON Output
All commands must support deterministic JSON output suitable for agents.

Example:
```bash
envdiff doctor . --json
```

JSON output must include:
- schema version
- command metadata
- inputs
- findings
- severity
- confidence where applicable
- locations
- summary counts
- explicit field names documented as part of the public CLI contract

---

## Data Model

## `EnvVarDefinition`
Fields:
- `name`
- `value`
- `normalized_value_kind`
- `file_path`
- `line_number`
- `source_type` (`dotenv`, `python`, `compose`)
- `is_duplicate`
- `parse_warnings`

## `EnvVarUsage`
Fields:
- `name`
- `file_path`
- `line_number`
- `usage_kind`
- `requiredness`
- `default_value`
- `source_type`

## `EnvVarContract`
Fields:
- `name`
- `definitions[]`
- `usages[]`
- `requiredness`
- `aliases[]`
- `secret_like`
- `placeholder_like`
- `status[]`

## `Finding`
Fields:
- `code`
- `severity`
- `title`
- `details`
- `variable_name`
- `locations[]`
- `related_variables[]`
- `suggested_fix`
- `confidence`
- `source_kind`
- `reason`
- `suppression_key`

---

## Output Design

### Human Output
Use Rich sections:
- Summary
- Missing
- Unused
- Alias candidates
- Secret hygiene
- Suggestions

### Agent Output
Deterministic JSON:
- stable ordering
- schema versioning
- explicit severity
- explicit location data
- machine-friendly status codes
- explicit confidence and reason fields for heuristic findings

---

## Severity Model

- `error`
- `warning`
- `info`

Examples:
- Missing required variable: `error`
- Missing optional variable: `warning`
- Unused variable: `info`
- Secret-like committed value: `warning` or `error` depending on mode

---

## Determinism Requirements

- Same inputs must yield identical findings and ordering.
- No non-deterministic sampling or LLM usage.
- File traversal order must be normalized.
- Similarity heuristics must use fixed thresholds.
- JSON output must be stable.

---

## CLI UX Requirements

- Helpful defaults
- Explicit exit codes
- `--json` for all major commands
- `--fail-on <severity>` support for CI
- `--quiet` and `--verbose` modes
- Clear parse warnings without crashing on malformed files
- Clear explanation of supported syntax patterns and unsupported constructs

---

## Exit Codes

Suggested:
- `0`: success, no failing findings
- `1`: execution/parsing error
- `2`: findings at or above `--fail-on` threshold

---

## Suggested v1 Repository Layout

```text
envdiff/
├── pyproject.toml
├── README.md
├── envdiff/
│   ├── cli.py
│   ├── models.py
│   ├── parsers/
│   │   ├── dotenv.py
│   │   ├── python_ast.py
│   │   └── compose.py
│   ├── analyzers/
│   │   ├── compare.py
│   │   ├── scan.py
│   │   ├── doctor.py
│   │   └── aliases.py
│   ├── render/
│   │   ├── human.py
│   │   └── json.py
│   └── utils/
│       ├── entropy.py
│       ├── normalize.py
│       └── paths.py
└── tests/
```

---

## Acceptance Criteria for v1

- `compare`, `scan`, and `doctor` implemented.
- Supports `.env`, Python env usage, and Docker Compose interpolation.
- Produces deterministic human and JSON output.
- Documents repo-local resolution behavior for monorepos and multi-service repos.
- Correctly detects missing, extra, unused, and alias-candidate variables.
- Flags secret-like values with heuristics.
- Includes tests for representative edge cases.
- Works offline on macOS/Linux with Python 3.11+.

---

## Future Expansion Areas

- GitHub Actions env parsing
- shell script support
- repo-local `.envrc` / direnv-style support using a conservative subset
- Pydantic settings inference
- `.env.example` generation
- monorepo/service grouping
- baseline/snapshot workflows for incremental adoption
- Kubernetes/Helm/Terraform support
- secret manager integration
