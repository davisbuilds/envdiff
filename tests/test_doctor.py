from __future__ import annotations

from envdiff.analyzers.doctor import doctor_repository
from envdiff.analyzers.scan import scan_repository


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
    assert "ENV005" in codes
    assert "ENV006" in codes
