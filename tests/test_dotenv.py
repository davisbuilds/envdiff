from __future__ import annotations

from envdiff.parsers.dotenv import parse_dotenv


def test_parse_dotenv_preserves_duplicates_and_types() -> None:
    result = parse_dotenv("tests/fixtures/dotenv/basic.env")

    names = [definition.name for definition in result.definitions]
    kinds = {definition.name: definition.normalized_value_kind for definition in result.definitions}

    assert names == ["DATABASE_URL", "DEBUG", "EMPTY", "QUOTED"]
    assert kinds["DATABASE_URL"] == "url"
    assert kinds["DEBUG"] == "boolean"
    assert kinds["EMPTY"] == "placeholder"
    assert kinds["QUOTED"] == "string"


def test_parse_dotenv_marks_duplicate_definitions() -> None:
    result = parse_dotenv("tests/fixtures/dotenv/duplicates.env")

    assert len(result.definitions) == 2
    assert result.definitions[0].is_duplicate is False
    assert result.definitions[1].is_duplicate is True


def test_parse_dotenv_records_warnings_for_unsupported_syntax() -> None:
    result = parse_dotenv("tests/fixtures/dotenv/malformed.env")

    assert len(result.warnings) == 1
    assert "unsupported dotenv syntax" in result.warnings[0]
