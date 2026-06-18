# Benchmark: Go CLI vs Python oracle (2026-06-18)

Wall-clock comparison of the Go `envdiff` against the retained Python oracle,
produced by `scripts/bench_go_vs_python.sh` (hyperfine 1.20).

## TL;DR

- **Startup dominates small-repo UX.** The compiled Go binary starts in ~1.9 ms
  vs ~72 ms for the Python console script — **~38× faster** per invocation
  before any analysis happens.
- **At scale Go is ~4.6× faster on `scan`** and **~9.7× faster on `doctor`**
  (1,000-file repos), staying roughly linear where the work is linear.
- **The `./envdiff` launcher measured here (`go run`) cost ~45 ms/run** — 24× a
  compiled binary — so most of Go's startup advantage was unrealized. *Fixed
  after this benchmark:* `./envdiff` now execs a cached compiled binary and only
  rebuilds when sources change.
- **`doctor` has an O(usages × defs) alias-detection cliff in both
  implementations** (Go 0.51 s → 12.7 s from 1k → 5k files). Scan concurrency
  does not help it; logged as a perf follow-up.

## Method

- **Headline comparison:** compiled Go binary (`go build -o … ./cmd/envdiff`) vs
  the Python console script in the project venv (`.venv/bin/envdiff`) — neither
  pays launcher overhead. A separate startup table also measures the as-shipped
  launchers (`./envdiff` → `go run`, `scripts/envdiff-python` → `uv run`).
- **Workload:** synthetic repos of N Python files (3 env references each: one
  required `os.environ`, one optional `os.getenv`, one with a default), half the
  `VAR`s defined in a top-level `.env`. `--json` output for both.
- **Tool:** hyperfine with warmup; run counts scale down with repo size.
- **Machine:** Apple M5 Pro, 15 cores, macOS 15 (darwin 25.4). Default
  `GOMAXPROCS` (all cores) for the concurrent scan.
- **Caveat — not purely apples-to-apples:** the Go Python scanner is regex-based
  while the Python oracle uses real `ast` parsing, so Go does *less* work on
  `.py` files by design. The parity gate guarantees equivalent *output*, not an
  equivalent algorithm. Treat `.py`-heavy ratios as a product comparison, not a
  language microbenchmark.

## Startup / launcher overhead

`scan` of a 1-file repo (so the time is almost entirely process startup):

| Invocation | Mean | vs Go binary |
|:--|--:|--:|
| `go binary` | 1.9 ms | 1.00× |
| `./envdiff` (`go run`) | 45.4 ms | 23.9× |
| `python venv` (`.venv/bin/envdiff`) | 71.8 ms | 37.7× |
| `scripts/envdiff-python` (`uv run`) | 85.9 ms | 45.1× |

`go run` rebuilds/links on every call; `uv run` adds ~14 ms of environment
resolution over the bare venv script.

## `scan` scaling (compiled binary vs venv script)

| Files | Go | Python | Speedup |
|--:|--:|--:|--:|
| 10 | 3.9 ms | 73.7 ms | 18.9× |
| 100 | 8.4 ms | 91.7 ms | 10.9× |
| 1,000 | 58.5 ms | 272.1 ms | 4.65× |
| 5,000 | 238.3 ms | 1,089.5 ms | 4.57× |

Both scale roughly linearly; Go settles around 4.6× at scale. The ratio shrinks
as repo size grows because fixed startup (where Go wins biggest) is amortized.

## `doctor` scaling

| Files | Go | Python | Speedup |
|--:|--:|--:|--:|
| 10 | 3.0 ms | 73.8 ms | 24.4× |
| 100 | 14.6 ms | 136.3 ms | 9.34× |
| 1,000 | 511.3 ms | 4,966 ms | 9.71× |
| 5,000 | ~12.7 s (spot) | aborted (> minutes) | — |

`doctor` is super-linear in both: Go goes 14.6 ms → 511 ms → ~12.7 s across
100 → 1,000 → 5,000 files (≈25× for each 5–10× of input). This fingerprints the
**O(usages × defs) alias-detection pass** in `DoctorRepository` (pairwise name
comparison), which the per-file scan concurrency does not touch. The 5,000-file
Python cell was not run to completion (same O(n²) shape on top of the
interpreter + AST cost); the Go spot value is from a single sweep.

## Takeaways

1. For the common case (small/medium repos, frequent invocation) the win is
   **startup**, not analysis — a compiled binary feels instant (~2 ms) where the
   Python CLI has a fixed ~72 ms tax.
2. **Ship a compiled binary.** The `go run` launcher threw away ~43 ms/run —
   since fixed: `./envdiff` execs a cached binary, rebuilding only on change.
3. **The next real perf target is `doctor`'s alias pass**, not scanning — see
   the perf follow-up in `docs/project/BACKLOG.md`.

## Reproduce

```bash
brew install hyperfine
uv sync --extra dev
scripts/bench_go_vs_python.sh                 # defaults: scan to 5k, doctor to 1k
BENCH_DOCTOR_SIZES="10 100 1000 5000" scripts/bench_go_vs_python.sh   # include the cliff
```
