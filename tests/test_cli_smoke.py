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
    assert "Summary:" in result.stdout
    assert "Errors:" in result.stdout
    assert "Warnings:" in result.stdout
    assert "Infos:" in result.stdout


def test_cli_doctor_write_baseline_smoke(tmp_path) -> None:
    baseline_path = tmp_path / "doctor-baseline.json"

    result = runner.invoke(
        app,
        [
            "doctor",
            "tests/fixtures/doctor/project",
            "--write-baseline",
            str(baseline_path),
        ],
    )

    assert result.exit_code == 0
    assert baseline_path.exists()
    assert "Baseline written:" in result.stdout


def test_cli_doctor_baseline_suppresses_findings(tmp_path) -> None:
    baseline_path = tmp_path / "doctor-baseline.json"
    first = runner.invoke(
        app,
        [
            "doctor",
            "tests/fixtures/doctor/project",
            "--write-baseline",
            str(baseline_path),
        ],
    )
    assert first.exit_code == 0

    second = runner.invoke(
        app,
        [
            "doctor",
            "tests/fixtures/doctor/project",
            "--baseline",
            str(baseline_path),
            "--fail-on",
            "warning",
        ],
    )

    assert second.exit_code == 0
    assert "Findings: 0" in second.stdout
    assert "No active findings." in second.stdout
