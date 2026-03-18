from __future__ import annotations

from src.parsers.compose import scan_compose_file


def test_scan_compose_file_detects_required_and_defaulted_values() -> None:
    result = scan_compose_file("tests/fixtures/compose/docker-compose.yml")

    assert [usage.name for usage in result.usages] == ["DATABASE_URL", "DEBUG"]
    assert result.usages[0].requiredness == "required"
    assert result.usages[1].requiredness == "optional_with_default"
    assert result.usages[1].default_value == "false"
