from __future__ import annotations

from src.parsers.python_ast import scan_python_file


def test_scan_python_file_detects_required_and_optional_usage() -> None:
    result = scan_python_file("tests/fixtures/python/sample_app.py")

    assert [usage.name for usage in result.usages] == ["DATABASE_URL", "DEBUG", "REDIS_URL"]
    assert result.usages[0].requiredness == "required"
    assert result.usages[1].requiredness == "optional_with_default"
    assert result.usages[1].default_value == "false"
    assert result.usages[2].requiredness == "optional"


def test_scan_python_file_ignores_dynamic_names() -> None:
    result = scan_python_file("tests/fixtures/python/unsupported.py")

    assert result.usages == ()
