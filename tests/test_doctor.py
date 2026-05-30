from __future__ import annotations

from collections import defaultdict

from src.analyzers.doctor import _skew_findings, doctor_repository
from src.analyzers.scan import scan_repository
from src.models import EnvVarDefinition, RepoScanResult, ResolutionDecision


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
        finding.variable_name for finding in findings if finding.code in {"ENV001", "ENV002"}
    }

    assert missing_names == {"API_KEY", "DEPLOY_ENV"}


def test_doctor_repository_reports_missing_usage_without_associated_env(tmp_path) -> None:
    app_file = tmp_path / "main.py"
    app_file.write_text('import os\nos.environ["API_KEY"]\n', encoding="utf-8")

    findings = doctor_repository(scan_repository(tmp_path))

    missing = next(finding for finding in findings if finding.code == "ENV001")
    assert missing.variable_name == "API_KEY"
    assert "no associated .env defines it" in missing.details


def test_doctor_repository_deduplicates_alias_findings_for_repeated_missing_usage(
    tmp_path,
) -> None:
    app_file = tmp_path / "main.py"
    (tmp_path / ".env").write_text("OPENAI_KEY=sk-test\n", encoding="utf-8")
    app_file.write_text(
        'import os\nos.environ["OPENAI_API_KEY"]\nos.environ["OPENAI_API_KEY"]\n',
        encoding="utf-8",
    )

    findings = doctor_repository(scan_repository(tmp_path))

    alias_findings = [finding for finding in findings if finding.code == "ENV007"]
    assert len(alias_findings) == 1
    assert alias_findings[0].related_variables == ("OPENAI_KEY",)


def test_template_skew_findings_skip_keys_already_seen() -> None:
    env_file = "/repo/.env"
    example_file = "/repo/.env.example"
    definition = EnvVarDefinition(
        name="STALE",
        value="",
        normalized_value_kind="placeholder",
        file_path=example_file,
        line_number=1,
        source_type="dotenv",
    )
    definitions_by_file = defaultdict(list, {env_file: [], example_file: [definition]})
    associated_usage_names = defaultdict(set)
    seen = {("ENV005", "STALE", example_file)}
    scan_result = RepoScanResult(
        root_path="/repo",
        resolutions=(
            ResolutionDecision(
                source_file="/repo/app.py",
                env_file=env_file,
                example_file=example_file,
            ),
        ),
    )

    findings = _skew_findings(
        scan_result,
        definitions_by_file,
        associated_usage_names,
        seen,
    )

    assert findings == []
