from __future__ import annotations

import re
from pathlib import Path

from envdiff.models import DotenvParseResult, EnvVarDefinition
from envdiff.utils.normalize import normalize_value_kind
from envdiff.utils.ordering import sort_definitions

KEY_VALUE_RE = re.compile(r"^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$")


def parse_dotenv(path: str | Path) -> DotenvParseResult:
    file_path = Path(path)
    seen: dict[str, int] = {}
    definitions: list[EnvVarDefinition] = []
    warnings: list[str] = []
    lines = file_path.read_text(encoding="utf-8").splitlines()

    for line_number, raw_line in enumerate(lines, start=1):
        stripped = raw_line.strip()
        if not stripped or stripped.startswith("#"):
            continue

        match = KEY_VALUE_RE.match(raw_line)
        if not match:
            warnings.append(f"{file_path}:{line_number}: unsupported dotenv syntax")
            continue

        name = match.group(1)
        value = _parse_value(match.group(2))
        duplicate_index = seen.get(name, 0)
        seen[name] = duplicate_index + 1

        definitions.append(
            EnvVarDefinition(
                name=name,
                value=value,
                normalized_value_kind=normalize_value_kind(value),
                file_path=str(file_path),
                line_number=line_number,
                source_type="dotenv",
                is_duplicate=duplicate_index > 0,
                parse_warnings=(),
            )
        )

    return DotenvParseResult(definitions=sort_definitions(definitions), warnings=tuple(warnings))


def _parse_value(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
        return value[1:-1]
    return value
