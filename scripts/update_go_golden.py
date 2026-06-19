#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[1]
GOLDEN_DIR = REPO_ROOT / "tests" / "golden" / "json"
REPO_TOKEN = "<REPO_ROOT>"
TMP_TOKEN = "<TMPDIR>"

_GO_BINARY: Path | None = None


def go_binary() -> Path:
    """Build the Go envdiff binary once and return its path.

    Go is the source of truth for the JSON contract, so goldens are generated
    from the compiled Go binary rather than the Python oracle.
    """
    global _GO_BINARY
    if _GO_BINARY is None:
        target = REPO_ROOT / "bin" / "envdiff"
        subprocess.run(
            ["go", "build", "-o", str(target), "./cmd/envdiff"],
            cwd=REPO_ROOT,
            check=True,
        )
        _GO_BINARY = target
    return _GO_BINARY


JSON_CASES: tuple[tuple[str, list[str], int], ...] = (
    (
        "compare-basic.json",
        [
            "compare",
            "tests/fixtures/compare/left.env",
            "tests/fixtures/compare/right.env",
            "--json",
        ],
        0,
    ),
    (
        "matrix-basic.json",
        [
            "matrix",
            "tests/fixtures/matrix/a.env",
            "tests/fixtures/matrix/b.env",
            "tests/fixtures/matrix/c.env",
            "--json",
        ],
        0,
    ),
    (
        "matrix-show-all.json",
        [
            "matrix",
            "tests/fixtures/matrix/a.env",
            "tests/fixtures/matrix/b.env",
            "tests/fixtures/matrix/c.env",
            "--show-all",
            "--json",
        ],
        0,
    ),
    ("scan-simple-repo.json", ["scan", "tests/fixtures/repos/simple_repo", "--json"], 0),
    ("scan-workflow-repo.json", ["scan", "tests/fixtures/repos/workflow_repo", "--json"], 0),
    ("scan-unicode-repo.json", ["scan", "tests/fixtures/repos/unicode_repo", "--json"], 0),
    ("generate-simple-repo.json", ["generate", "tests/fixtures/repos/simple_repo", "--json"], 0),
    (
        "generate-simple-repo-annotated.json",
        ["generate", "tests/fixtures/repos/simple_repo", "--annotate", "--json"],
        0,
    ),
    ("doctor-project.json", ["doctor", "tests/fixtures/doctor/project", "--json"], 2),
)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Regenerate JSON goldens from the Go envdiff binary (the contract source)."
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Verify committed goldens match the Go binary without rewriting files.",
    )
    args = parser.parse_args()

    failures: list[str] = []
    for filename, command_args, expected_code in JSON_CASES:
        payload = run_json_case(command_args, expected_code=expected_code)
        rendered = render_json(normalize_payload(payload))
        target = GOLDEN_DIR / filename
        check_or_write(target, rendered, check=args.check, failures=failures)

    with tempfile.TemporaryDirectory(prefix="envdiff-go-golden-") as tmp:
        tmp_path = Path(tmp)
        baseline_path = tmp_path / "doctor-baseline.json"
        run_command(
            ["doctor", "tests/fixtures/doctor/project", "--write-baseline", str(baseline_path)],
            expected_code=0,
        )
        payload = run_json_case(
            [
                "doctor",
                "tests/fixtures/doctor/project",
                "--baseline",
                str(baseline_path),
                "--json",
            ],
            expected_code=0,
        )
        rendered = render_json(normalize_payload(payload, tmp_path=tmp_path))
        check_or_write(
            GOLDEN_DIR / "doctor-project-baseline.json",
            rendered,
            check=args.check,
            failures=failures,
        )

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    return 0


def run_json_case(command_args: Sequence[str], *, expected_code: int) -> Any:
    result = run_command(command_args, expected_code=expected_code)
    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"{' '.join(command_args)} did not emit valid JSON: {exc}") from exc


def run_command(
    command_args: Sequence[str],
    *,
    expected_code: int,
) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        [str(go_binary()), *command_args],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != expected_code:
        raise SystemExit(
            f"{' '.join(command_args)} exited {result.returncode}, expected {expected_code}\n"
            f"stdout:\n{result.stdout}\n"
            f"stderr:\n{result.stderr}"
        )
    return result


def normalize_payload(payload: Any, *, tmp_path: Path | None = None) -> Any:
    replacements = path_replacements(REPO_ROOT, REPO_TOKEN)
    if tmp_path is not None:
        replacements.update(path_replacements(tmp_path, TMP_TOKEN))
    return normalize_value(payload, replacements)


def path_replacements(path: Path, token: str) -> dict[str, str]:
    return {
        str(path.absolute()).replace("\\", "/"): token,
        str(path.resolve()).replace("\\", "/"): token,
    }


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


def render_json(payload: Any) -> str:
    return json.dumps(payload, indent=2, sort_keys=True, ensure_ascii=False) + "\n"


def check_or_write(target: Path, rendered: str, *, check: bool, failures: list[str]) -> None:
    if check:
        existing = target.read_text(encoding="utf-8") if target.exists() else None
        if existing != rendered:
            failures.append(f"{target.relative_to(REPO_ROOT)} is not up to date")
        return

    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(rendered, encoding="utf-8")


if __name__ == "__main__":
    raise SystemExit(main())
