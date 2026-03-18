from __future__ import annotations

from pathlib import Path

from src.parsers.dotenv import parse_dotenv


def matrix_dotenv_files(
    paths: tuple[str, ...] | list[str],
    *,
    show_all: bool = False,
) -> dict[str, object]:
    normalized_paths = tuple(str(Path(path)) for path in paths)
    parsed = [(path, parse_dotenv(path)) for path in normalized_paths]
    latest_by_path = {path: _latest_by_name(result.definitions) for path, result in parsed}
    definitions_by_path = {path: result.definitions for path, result in parsed}
    all_names = sorted({name for definitions in latest_by_path.values() for name in definitions})

    variables = []
    inconsistent_count = 0
    for name in all_names:
        files = []
        kinds = set()
        present_in = []
        missing_in = []
        duplicates_in = []

        for path in normalized_paths:
            definition = latest_by_path[path].get(name)
            if definition is None:
                files.append({"path": path, "presence": "missing"})
                missing_in.append(path)
                continue

            kinds.add(definition.normalized_value_kind)
            present_in.append(path)
            if any(
                candidate.name == name and candidate.is_duplicate
                for candidate in definitions_by_path[path]
            ):
                duplicates_in.append(path)
            files.append(
                {
                    "path": path,
                    "presence": "present",
                    "value_kind": definition.normalized_value_kind,
                    "is_duplicate": path in duplicates_in,
                }
            )

        inconsistent = (
            len(missing_in) > 0 or len(kinds) > 1 or len(duplicates_in) > 0
        )
        if inconsistent:
            inconsistent_count += 1

        variable = {
            "name": name,
            "status": "inconsistent" if inconsistent else "consistent",
            "present_in": present_in,
            "missing_in": missing_in,
            "duplicates_in": duplicates_in,
            "value_kinds": sorted(kinds),
            "files": files,
        }
        if show_all or inconsistent:
            variables.append(variable)

    warnings = sorted(
        warning for _, result in parsed for warning in result.warnings
    )
    return {
        "paths": list(normalized_paths),
        "show_all": show_all,
        "file_count": len(normalized_paths),
        "variable_count": len(all_names),
        "inconsistent_variable_count": inconsistent_count,
        "variables": variables,
        "warnings": warnings,
    }


def _latest_by_name(definitions) -> dict[str, object]:
    latest = {}
    for definition in definitions:
        latest[definition.name] = definition
    return latest
