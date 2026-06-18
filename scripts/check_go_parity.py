#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
import sys
import tempfile
from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[1]
REPO_TOKEN = "<REPO_ROOT>"
TMP_TOKEN = "<TMPDIR>"

JSON_CASES: tuple[tuple[str, list[str]], ...] = (
    (
        "compare",
        [
            "compare",
            "tests/fixtures/compare/left.env",
            "tests/fixtures/compare/right.env",
            "--json",
        ],
    ),
    (
        "matrix",
        [
            "matrix",
            "tests/fixtures/matrix/a.env",
            "tests/fixtures/matrix/b.env",
            "tests/fixtures/matrix/c.env",
            "--json",
        ],
    ),
    (
        "matrix-show-all",
        [
            "matrix",
            "tests/fixtures/matrix/a.env",
            "tests/fixtures/matrix/b.env",
            "tests/fixtures/matrix/c.env",
            "--show-all",
            "--json",
        ],
    ),
    ("scan-simple", ["scan", "tests/fixtures/repos/simple_repo", "--json"]),
    ("scan-workflow", ["scan", "tests/fixtures/repos/workflow_repo", "--json"]),
    ("generate", ["generate", "tests/fixtures/repos/simple_repo", "--json"]),
    (
        "generate-annotated",
        ["generate", "tests/fixtures/repos/simple_repo", "--annotate", "--json"],
    ),
    ("doctor", ["doctor", "tests/fixtures/doctor/project", "--json"]),
)

HUMAN_CASES: tuple[tuple[str, list[str], tuple[str, ...]], ...] = (
    (
        "compare",
        ["compare", "tests/fixtures/compare/left.env", "tests/fixtures/compare/right.env"],
        ("Missing in left",),
    ),
    ("scan", ["scan", "tests/fixtures/repos/simple_repo"], ("Contracts: 3",)),
    (
        "matrix",
        [
            "matrix",
            "tests/fixtures/matrix/a.env",
            "tests/fixtures/matrix/b.env",
            "tests/fixtures/matrix/c.env",
        ],
        ("Matrix files: 3",),
    ),
    (
        "generate",
        ["generate", "tests/fixtures/repos/simple_repo"],
        ("DATABASE_URL=", "REDIS_URL=", "DEBUG="),
    ),
    (
        "doctor",
        ["doctor", "tests/fixtures/doctor/project", "--fail-on", "warning"],
        ("Summary:", "ENV001"),
    ),
)

EXIT_CASES: tuple[tuple[str, list[str]], ...] = (
    ("matrix-single", ["matrix", "tests/fixtures/matrix/a.env"]),
    ("doctor-threshold", ["doctor", "tests/fixtures/doctor/project", "--fail-on", "warning"]),
    ("doctor-invalid-fail-on", ["doctor", "tests/fixtures/doctor/project", "--fail-on", "debug"]),
    ("generate-drift", ["generate", "tests/fixtures/repos/simple_repo", "--check"]),
)


def main() -> int:
    parser = argparse.ArgumentParser(description="Check Python and Go envdiff parity.")
    parser.parse_args()

    go_binary = shutil.which("go")
    if go_binary is None:
        print("go is not available on PATH", file=sys.stderr)
        return 1

    failures: list[str] = []
    for name, command_args in JSON_CASES:
        compare_json_case(name, command_args, failures)

    compare_baseline_json_case(failures)

    for name, command_args, smoke_strings in HUMAN_CASES:
        compare_human_case(name, command_args, smoke_strings, failures)

    for name, command_args in EXIT_CASES:
        compare_exit_case(name, command_args, failures)

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    print("Python and Go parity checks passed.")
    return 0


def compare_json_case(name: str, command_args: Sequence[str], failures: list[str]) -> None:
    python = run_python(command_args)
    go = run_go(command_args)
    if normalize_json_output(python.stdout) != normalize_json_output(go.stdout):
        failures.append(f"{name}: JSON output differs")
    if normalized_exit_code(go) != python.returncode:
        failures.append(
            f"{name}: exit code differs, Python={python.returncode} Go={normalized_exit_code(go)}"
        )


def compare_baseline_json_case(failures: list[str]) -> None:
    with tempfile.TemporaryDirectory(prefix="envdiff-parity-") as tmp:
        tmp_path = Path(tmp)
        baseline = tmp_path / "doctor-baseline.json"

        run_python(
            [
                "doctor",
                "tests/fixtures/doctor/project",
                "--write-baseline",
                str(baseline),
            ]
        )

        python = run_python(
            [
                "doctor",
                "tests/fixtures/doctor/project",
                "--baseline",
                str(baseline),
                "--json",
            ]
        )
        go = run_go(
            [
                "doctor",
                "tests/fixtures/doctor/project",
                "--baseline",
                str(baseline),
                "--json",
            ]
        )

        replacements = path_replacements(REPO_ROOT, REPO_TOKEN)
        replacements.update(path_replacements(tmp_path, TMP_TOKEN))
        if normalize_json_output(python.stdout, replacements) != normalize_json_output(
            go.stdout,
            replacements,
        ):
            failures.append("doctor-baseline: JSON output differs")
        if normalized_exit_code(go) != python.returncode:
            failures.append(
                "doctor-baseline: exit code differs, "
                f"Python={python.returncode} Go={normalized_exit_code(go)}"
            )


def compare_human_case(
    name: str,
    command_args: Sequence[str],
    smoke_strings: Sequence[str],
    failures: list[str],
) -> None:
    python = run_python(command_args)
    go = run_go(command_args)
    if normalized_exit_code(go) != python.returncode:
        failures.append(
            f"{name}: human exit code differs, Python={python.returncode} "
            f"Go={normalized_exit_code(go)}"
        )
    for smoke in smoke_strings:
        if smoke not in python.stdout:
            failures.append(f"{name}: Python human output missing {smoke!r}")
        if smoke not in go.stdout:
            failures.append(f"{name}: Go human output missing {smoke!r}")


def compare_exit_case(name: str, command_args: Sequence[str], failures: list[str]) -> None:
    python = run_python(command_args)
    go = run_go(command_args)
    if normalized_exit_code(go) != python.returncode:
        failures.append(
            f"{name}: exit code differs, Python={python.returncode} Go={normalized_exit_code(go)}"
        )


def run_python(command_args: Sequence[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [str(REPO_ROOT / "envdiff"), *command_args],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )


def run_go(command_args: Sequence[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["go", "run", "./cmd/envdiff", *command_args],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )


def normalized_exit_code(result: subprocess.CompletedProcess[str]) -> int:
    if result.returncode == 1:
        match = re.search(r"(?:^|\n)exit status (\d+)(?:\n|$)", result.stderr)
        if match:
            return int(match.group(1))
    return result.returncode


def normalize_json_output(
    raw_json: str,
    replacements: Mapping[str, str] | None = None,
) -> Any:
    if replacements is None:
        replacements = path_replacements(REPO_ROOT, REPO_TOKEN)
    return normalize_value(json.loads(raw_json), replacements)


def normalize_value(value: Any, replacements: Mapping[str, str]) -> Any:
    if isinstance(value, dict):
        return {key: normalize_value(item, replacements) for key, item in value.items()}
    if isinstance(value, list):
        return [normalize_value(item, replacements) for item in value]
    if isinstance(value, str):
        normalized = value.replace("\\", "/")
        for source, replacement in replacements.items():
            normalized = normalized.replace(source, replacement)
        return normalized
    return value


def path_replacements(path: Path, token: str) -> dict[str, str]:
    return {
        str(path.absolute()).replace("\\", "/"): token,
        str(path.resolve()).replace("\\", "/"): token,
    }


if __name__ == "__main__":
    raise SystemExit(main())
