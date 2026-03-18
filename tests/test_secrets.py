from __future__ import annotations

from src.analyzers.scan import scan_repository
from src.analyzers.secrets import secret_and_placeholder_findings


def test_secret_and_placeholder_findings_detect_expected_values() -> None:
    scan_result = scan_repository("tests/fixtures/doctor/project")
    findings = secret_and_placeholder_findings(scan_result)
    codes = {finding.code for finding in findings}

    assert "ENV008" in codes
    assert "ENV009" in codes
