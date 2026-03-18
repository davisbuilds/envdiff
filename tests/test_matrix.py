from __future__ import annotations

from src.analyzers.matrix import matrix_dotenv_files

FIXTURE_PATHS = (
    "tests/fixtures/matrix/a.env",
    "tests/fixtures/matrix/b.env",
    "tests/fixtures/matrix/c.env",
)


def test_matrix_reports_only_inconsistent_variables_by_default() -> None:
    result = matrix_dotenv_files(FIXTURE_PATHS)

    names = [variable["name"] for variable in result["variables"]]

    assert result["file_count"] == 3
    assert result["variable_count"] == 4
    assert result["inconsistent_variable_count"] == 3
    assert names == ["API_KEY", "DATABASE_URL", "DEBUG"]


def test_matrix_tracks_missing_kind_and_duplicate_signals() -> None:
    result = matrix_dotenv_files(FIXTURE_PATHS)
    variables = {variable["name"]: variable for variable in result["variables"]}

    api_key = variables["API_KEY"]
    debug = variables["DEBUG"]
    database_url = variables["DATABASE_URL"]

    assert api_key["missing_in"] == [FIXTURE_PATHS[0], FIXTURE_PATHS[2]]
    assert api_key["value_kinds"] == ["secret_like"]
    assert database_url["value_kinds"] == ["placeholder", "url"]
    assert debug["duplicates_in"] == [FIXTURE_PATHS[2]]


def test_matrix_show_all_includes_consistent_variables() -> None:
    result = matrix_dotenv_files(FIXTURE_PATHS, show_all=True)

    names = [variable["name"] for variable in result["variables"]]

    assert names == ["API_KEY", "DATABASE_URL", "DEBUG", "SHARED_MODE"]
