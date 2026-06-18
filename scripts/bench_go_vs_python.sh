#!/usr/bin/env bash
# Benchmark the Go envdiff CLI against the Python oracle.
#
# Headline comparison is the compiled Go binary vs the Python console script in
# the project venv (no go-run / uv-run launcher overhead). A separate startup
# table also measures the as-shipped launchers (./envdiff -> go run,
# scripts/envdiff-python -> uv run) so their per-invocation cost is visible.
#
# scan is swept further than doctor: doctor's alias pass is O(usages x defs),
# so very large repos are a documented scaling cliff rather than a routine cell.
#
# Usage: scripts/bench_go_vs_python.sh [output_dir]
# Env:   BENCH_SCAN_SIZES, BENCH_DOCTOR_SIZES (space-separated file counts)
# Requires: hyperfine, go, python3, and a synced venv (.venv/bin/envdiff).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

OUT_DIR="${1:-$(mktemp -d -t envdiff-bench)}"
mkdir -p "$OUT_DIR"

SCAN_SIZES="${BENCH_SCAN_SIZES:-10 100 1000 5000}"
DOCTOR_SIZES="${BENCH_DOCTOR_SIZES:-10 100 1000}"

for tool in hyperfine go python3; do
	command -v "$tool" >/dev/null || { echo "missing required tool: $tool" >&2; exit 1; }
done

PY_BIN="$REPO_ROOT/.venv/bin/envdiff"
[ -x "$PY_BIN" ] || { echo "missing $PY_BIN (run: uv sync --extra dev)" >&2; exit 1; }

WORK="$(mktemp -d -t envdiff-bench-work)"
trap 'rm -rf "$WORK"' EXIT

GO_BIN="$WORK/envdiff"
echo "building Go binary..."
go build -o "$GO_BIN" ./cmd/envdiff

# Fewer repetitions as repos (and per-run cost) grow, so the sweep stays quick.
runs_for() { if [ "$1" -le 100 ]; then echo 10; elif [ "$1" -le 1000 ]; then echo 6; else echo 3; fi; }
warmup_for() { if [ "$1" -le 1000 ]; then echo 2; else echo 1; fi; }

# Generate (once per size) a synthetic repo of N python files, half their VARs
# defined in .env.
ensure_repo() {
	local root="$WORK/repo-$1"
	[ -d "$root" ] && { echo "$root"; return; }
	python3 - "$root" "$1" <<'PY'
import os, sys
root, n = sys.argv[1], int(sys.argv[2])
os.makedirs(root, exist_ok=True)
with open(os.path.join(root, ".env"), "w") as f:
    for i in range(0, n, 2):
        f.write(f"VAR{i}=value{i}\n")
for i in range(n):
    d = os.path.join(root, f"pkg{i % 50:03d}")
    os.makedirs(d, exist_ok=True)
    with open(os.path.join(d, f"mod{i:05d}.py"), "w") as f:
        f.write("import os\n\n")
        f.write(f'A = os.getenv("VAR{i}")\n')
        f.write(f'B = os.environ["OTHER{i}"]\n')
        f.write(f'C = os.getenv("THIRD{i}", "default")\n')
PY
	echo "$root"
}

bench_cmd() { # cmd size [extra hyperfine flags...]
	local cmd="$1" size="$2"; shift 2
	local repo; repo="$(ensure_repo "$size")"
	echo ">> $cmd @ $size files"
	hyperfine -N --warmup "$(warmup_for "$size")" --runs "$(runs_for "$size")" "$@" \
		--export-markdown "$OUT_DIR/$cmd-$size.md" \
		--export-json "$OUT_DIR/$cmd-$size.json" \
		-n "go ($size)" "$GO_BIN $cmd $repo --json" \
		-n "python ($size)" "$PY_BIN $cmd $repo --json"
}

echo "results -> $OUT_DIR"

for size in $SCAN_SIZES; do
	bench_cmd scan "$size"
done
for size in $DOCTOR_SIZES; do
	# doctor exits 2 when findings meet --fail-on; expected on the synthetic repo.
	bench_cmd doctor "$size" --ignore-failure
done

# --- Startup / launcher overhead: scan of a 1-file repo ---------------------
tiny="$(ensure_repo 1)"
echo ">> startup + launcher overhead (scan, 1 file)"
hyperfine -N --warmup 3 --runs 15 \
	--export-markdown "$OUT_DIR/startup.md" \
	--export-json "$OUT_DIR/startup.json" \
	-n "go binary" "$GO_BIN scan $tiny --json" \
	-n "python venv" "$PY_BIN scan $tiny --json" \
	-n "./envdiff (go run)" "$REPO_ROOT/envdiff scan $tiny --json" \
	-n "envdiff-python (uv run)" "$REPO_ROOT/scripts/envdiff-python scan $tiny --json"

echo "done. markdown + json in $OUT_DIR"
