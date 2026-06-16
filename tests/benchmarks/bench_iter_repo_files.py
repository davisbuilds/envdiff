from __future__ import annotations

import argparse
import shutil
import statistics
import tempfile
import time
from pathlib import Path

from envdiff.utils.paths import iter_repo_files


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Benchmark repository file iteration with ignored directories."
    )
    parser.add_argument("--rounds", type=int, default=7)
    parser.add_argument("--ignored-files", type=int, default=2500)
    parser.add_argument("--tracked-files", type=int, default=250)
    args = parser.parse_args()

    root = Path(tempfile.mkdtemp(prefix="envdiff-bench-iter-"))
    try:
        _build_fixture(root, ignored_files=args.ignored_files, tracked_files=args.tracked_files)
        timings = []
        file_count = 0
        for _ in range(args.rounds):
            start = time.perf_counter()
            files = iter_repo_files(root)
            timings.append(time.perf_counter() - start)
            file_count = len(files)

        mean_ms = statistics.fmean(timings) * 1000
        min_ms = min(timings) * 1000
        max_ms = max(timings) * 1000
        print(f"metric=iter_repo_files_mean_ms value={mean_ms:.3f} lower_is_better=true")
        print(f"metric=iter_repo_files_min_ms value={min_ms:.3f} lower_is_better=true")
        print(f"metric=iter_repo_files_max_ms value={max_ms:.3f} lower_is_better=true")
        print(
            f"rounds={args.rounds} "
            f"ignored_files={args.ignored_files} "
            f"tracked_files={args.tracked_files}"
        )
        print(f"returned_files={file_count}")
    finally:
        shutil.rmtree(root)


def _build_fixture(root: Path, *, ignored_files: int, tracked_files: int) -> None:
    ignored_root = root / "node_modules" / "package"
    tracked_root = root / "src"
    ignored_root.mkdir(parents=True)
    tracked_root.mkdir(parents=True)

    for index in range(ignored_files):
        (ignored_root / f"ignored_{index}.js").write_text("process.env.IGNORED\n", encoding="utf-8")

    for index in range(tracked_files):
        (tracked_root / f"tracked_{index}.py").write_text(
            "import os\nos.getenv('APP_ENV')\n",
            encoding="utf-8",
        )

    (root / ".env").write_text("APP_ENV=dev\n", encoding="utf-8")


if __name__ == "__main__":
    main()
