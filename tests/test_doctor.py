from __future__ import annotations

from src.analyzers.doctor import doctor_repository
from src.analyzers.scan import scan_repository


def test_doctor_repository_emits_core_findings() -> None:
    scan_result = scan_repository("tests/fixtures/doctor/project")
    findings = doctor_repository(scan_result)
    codes = [finding.code for finding in findings]

    assert "ENV001" in codes
    assert "ENV002" in codes
    assert "ENV003" in codes
    assert "ENV007" in codes
    assert "ENV008" in codes
    assert "ENV009" in codes
    assert "ENV006" in codes


def test_doctor_repository_does_not_emit_unused_for_env_example_entries() -> None:
    scan_result = scan_repository("tests/fixtures/doctor/project")
    findings = doctor_repository(scan_result)

    example_unused = [
        finding
        for finding in findings
        if finding.code == "ENV003"
        and any(location.file_path.endswith(".env.example") for location in finding.locations)
    ]

    assert example_unused == []


def test_doctor_repository_emits_template_skew_for_stale_example_only_variable(
    tmp_path,
) -> None:
    project = tmp_path / "project"
    app_dir = project / "app"
    app_dir.mkdir(parents=True)
    (project / ".env").write_text("DATABASE_URL=postgres://localhost/db\n", encoding="utf-8")
    (project / ".env.example").write_text(
        "DATABASE_URL=\nLEGACY_TEMPLATE=\n",
        encoding="utf-8",
    )
    (app_dir / "main.py").write_text(
        'import os\n\ndatabase_url = os.environ["DATABASE_URL"]\n',
        encoding="utf-8",
    )

    findings = doctor_repository(scan_repository(project))

    template_skew = [finding for finding in findings if finding.code == "ENV005"]

    assert len(template_skew) == 1
    assert template_skew[0].severity == "info"
    assert template_skew[0].variable_name == "LEGACY_TEMPLATE"


def test_doctor_repository_reports_missing_workflow_secret_references() -> None:
    findings = doctor_repository(scan_repository("tests/fixtures/repos/workflow_repo"))

    missing_names = {
        finding.variable_name
        for finding in findings
        if finding.code in {"ENV001", "ENV002"}
    }

    assert missing_names == {"API_KEY", "DEPLOY_ENV"}
