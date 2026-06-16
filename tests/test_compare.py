from __future__ import annotations

from envdiff.analyzers.compare import compare_dotenv_files


def test_compare_dotenv_files_reports_missing_duplicates_and_kind_changes() -> None:
    result = compare_dotenv_files(
        "tests/fixtures/compare/left.env",
        "tests/fixtures/compare/right.env",
    )

    assert result["missing_in_left"] == ["FEATURE"]
    assert result["missing_in_right"] == ["DUP_KEY"]
    assert result["duplicates_in_left"] == ["DUP_KEY"]
    assert result["duplicates_in_right"] == []
    assert result["differing_values"][0]["name"] == "DATABASE_URL"
    assert result["differing_values"][0]["left_kind"] == "url"
    assert result["differing_values"][0]["right_kind"] == "integer"
