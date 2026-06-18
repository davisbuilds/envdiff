from __future__ import annotations

import shutil
import subprocess

import pytest


def test_go_parity_script() -> None:
    if shutil.which("go") is None:
        pytest.skip("go is not available on PATH")

    result = subprocess.run(
        ["uv", "run", "python", "scripts/check_go_parity.py"],
        capture_output=True,
        text=True,
        check=False,
    )

    assert result.returncode == 0, result.stdout + result.stderr
