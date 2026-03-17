from __future__ import annotations

from collections import defaultdict
from pathlib import Path

from envdiff.models import EnvVarContract, RepoScanResult, ResolutionDecision
from envdiff.parsers.compose import scan_compose_file
from envdiff.parsers.dotenv import parse_dotenv
from envdiff.parsers.python_ast import scan_python_file
from envdiff.utils.ordering import sort_contracts, sort_definitions, sort_usages
from envdiff.utils.paths import find_nearest_named_file, iter_repo_files

COMPOSE_FILENAMES = {"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}


def scan_repository(path: str | Path) -> RepoScanResult:
    root = Path(path).resolve()
    definitions = []
    usages = []
    warnings: list[str] = []
    resolution_map: dict[str, ResolutionDecision] = {}

    for file_path in iter_repo_files(root):
        if file_path.name in {".env", ".env.example"}:
            result = parse_dotenv(file_path)
            definitions.extend(result.definitions)
            warnings.extend(result.warnings)
            continue

        if file_path.suffix == ".py":
            result = scan_python_file(file_path)
            usages.extend(result.usages)
            warnings.extend(result.warnings)
            resolution_map[str(file_path)] = _resolve_usage_file(file_path, root)
            continue

        if file_path.name in COMPOSE_FILENAMES:
            result = scan_compose_file(file_path)
            usages.extend(result.usages)
            warnings.extend(result.warnings)
            resolution_map[str(file_path)] = _resolve_usage_file(file_path, root)

    contracts = _build_contracts(definitions, usages, resolution_map)
    resolutions = tuple(sorted(resolution_map.values(), key=lambda decision: decision.source_file))

    return RepoScanResult(
        root_path=str(root),
        definitions=sort_definitions(definitions),
        usages=sort_usages(usages),
        contracts=contracts,
        resolutions=resolutions,
        warnings=tuple(sorted(warnings)),
    )


def _resolve_usage_file(file_path: Path, root: Path) -> ResolutionDecision:
    env_file = find_nearest_named_file(file_path, root, ".env")
    example_file = find_nearest_named_file(file_path, root, ".env.example")
    notes = []

    if env_file:
        notes.append(f"env:{env_file}")
    if example_file:
        notes.append(f"example:{example_file}")
    if not notes:
        notes.append("no associated dotenv files found")

    return ResolutionDecision(
        source_file=str(file_path),
        env_file=str(env_file) if env_file else None,
        example_file=str(example_file) if example_file else None,
        notes=tuple(notes),
    )


def _build_contracts(
    definitions,
    usages,
    resolution_map: dict[str, ResolutionDecision],
) -> tuple[EnvVarContract, ...]:
    by_name: dict[str, dict[str, list]] = defaultdict(
        lambda: {"definitions": [], "usages": [], "notes": []}
    )

    for definition in definitions:
        by_name[definition.name]["definitions"].append(definition)

    for usage in usages:
        by_name[usage.name]["usages"].append(usage)
        resolution = resolution_map.get(usage.file_path)
        if resolution:
            by_name[usage.name]["notes"].extend(resolution.notes)

    contracts = []
    for name, payload in by_name.items():
        requiredness = _infer_requiredness(payload["usages"])
        statuses = []
        if payload["usages"]:
            statuses.append("referenced")
        if payload["definitions"]:
            statuses.append("defined")
        if payload["usages"] and not payload["definitions"]:
            statuses.append("undefined")
        if payload["definitions"] and not payload["usages"]:
            statuses.append("unreferenced")

        contracts.append(
            EnvVarContract(
                name=name,
                definitions=sort_definitions(payload["definitions"]),
                usages=sort_usages(payload["usages"]),
                requiredness=requiredness,
                status=tuple(sorted(statuses)),
                resolution_notes=tuple(sorted(set(payload["notes"]))),
            )
        )

    return sort_contracts(contracts)


def _infer_requiredness(usages) -> str:
    requirednesses = {usage.requiredness for usage in usages}
    if "required" in requirednesses:
        return "required"
    if "optional_with_default" in requirednesses:
        return "optional_with_default"
    if "optional" in requirednesses:
        return "optional"
    return "unknown"
