from __future__ import annotations

from envdiff.analyzers.scan import scan_repository


def test_scan_repository_builds_contracts_for_simple_repo() -> None:
    result = scan_repository("tests/fixtures/repos/simple_repo")

    contract_names = [contract.name for contract in result.contracts]

    assert contract_names == ["DATABASE_URL", "DEBUG", "REDIS_URL"]
    assert len(result.resolutions) == 1
    assert result.resolutions[0].env_file.endswith("tests/fixtures/repos/simple_repo/.env")
    assert result.resolutions[0].example_file.endswith(
        "tests/fixtures/repos/simple_repo/.env.example"
    )


def test_scan_repository_uses_nearest_env_files_in_monorepo() -> None:
    result = scan_repository("tests/fixtures/repos/monorepo")

    resolutions = {resolution.source_file: resolution for resolution in result.resolutions}

    api_resolution = next(
        resolution for path, resolution in resolutions.items() if path.endswith("apps/api/app.py")
    )
    web_resolution = next(
        resolution for path, resolution in resolutions.items() if path.endswith("apps/web/app.py")
    )

    assert api_resolution.env_file.endswith("apps/api/.env")
    assert web_resolution.env_file.endswith("apps/web/.env")
