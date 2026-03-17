from __future__ import annotations

from pathlib import Path

from envdiff.parsers.dotenv import parse_dotenv


def compare_dotenv_files(left: str | Path, right: str | Path) -> dict[str, object]:
    left_result = parse_dotenv(left)
    right_result = parse_dotenv(right)

    left_latest = _latest_by_name(left_result.definitions)
    right_latest = _latest_by_name(right_result.definitions)

    left_names = set(left_latest)
    right_names = set(right_latest)
    shared_names = sorted(left_names & right_names)

    differing = []
    for name in shared_names:
        left_definition = left_latest[name]
        right_definition = right_latest[name]
        if (
            left_definition.normalized_value_kind != right_definition.normalized_value_kind
            or left_definition.value != right_definition.value
        ):
            differing.append(
                {
                    "name": name,
                    "left_value": left_definition.value,
                    "right_value": right_definition.value,
                    "left_kind": left_definition.normalized_value_kind,
                    "right_kind": right_definition.normalized_value_kind,
                }
            )

    return {
        "left_path": str(Path(left)),
        "right_path": str(Path(right)),
        "missing_in_left": sorted(right_names - left_names),
        "missing_in_right": sorted(left_names - right_names),
        "duplicates_in_left": sorted(
            definition.name for definition in left_result.definitions if definition.is_duplicate
        ),
        "duplicates_in_right": sorted(
            definition.name for definition in right_result.definitions if definition.is_duplicate
        ),
        "differing_values": differing,
        "warnings": sorted((*left_result.warnings, *right_result.warnings)),
    }


def _latest_by_name(definitions) -> dict[str, object]:
    latest = {}
    for definition in definitions:
        latest[definition.name] = definition
    return latest

