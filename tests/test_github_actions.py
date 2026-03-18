from __future__ import annotations

from src.parsers.github_actions import scan_github_actions_file


def test_scan_github_actions_file_extracts_secret_and_var_references() -> None:
    result = scan_github_actions_file("tests/fixtures/github_actions/deploy.yml")

    usages = {(usage.name, usage.requiredness, usage.default_value) for usage in result.usages}

    assert usages == {
        ("API_KEY", "required", None),
        ("DATABASE_URL", "required", None),
        ("DEPLOY_ENV", "optional_with_default", "staging"),
    }
    assert all(usage.source_type == "github_actions" for usage in result.usages)
