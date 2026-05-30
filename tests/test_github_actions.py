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


def test_scan_github_actions_file_handles_blank_and_unquoted_defaults(tmp_path) -> None:
    workflow = tmp_path / "deploy.yml"
    workflow.write_text(
        "\n".join(
            [
                "env:",
                "  EMPTY: ${{ vars.EMPTY || }}",
                "  REGION: ${{ vars.REGION || us-east-1 }}",
            ]
        ),
        encoding="utf-8",
    )

    result = scan_github_actions_file(workflow)

    usages = {usage.name: usage for usage in result.usages}
    assert usages["EMPTY"].requiredness == "required"
    assert usages["EMPTY"].default_value is None
    assert usages["REGION"].requiredness == "optional_with_default"
    assert usages["REGION"].default_value == "us-east-1"
