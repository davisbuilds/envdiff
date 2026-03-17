from __future__ import annotations

from typer.testing import CliRunner

from envdiff.cli import app

runner = CliRunner()


def test_cli_help_smoke() -> None:
    result = runner.invoke(app, ["--help"])

    assert result.exit_code == 0
    assert "compare" in result.stdout
    assert "scan" in result.stdout
    assert "doctor" in result.stdout


def test_cli_compare_json_smoke() -> None:
    result = runner.invoke(
        app,
        [
            "compare",
            "tests/fixtures/compare/left.env",
            "tests/fixtures/compare/right.env",
            "--json",
        ],
    )

    assert result.exit_code == 0
    assert '"command": "compare"' in result.stdout
    assert '"missing_in_left"' in result.stdout


def test_cli_scan_json_smoke() -> None:
    result = runner.invoke(
        app,
        ["scan", "tests/fixtures/repos/simple_repo", "--json"],
    )

    assert result.exit_code == 0
    assert '"command": "scan"' in result.stdout
    assert '"contracts"' in result.stdout


def test_cli_doctor_threshold_smoke() -> None:
    result = runner.invoke(
        app,
        ["doctor", "tests/fixtures/doctor/project", "--fail-on", "warning"],
    )

    assert result.exit_code == 2
    assert "ENV001" in result.stdout
