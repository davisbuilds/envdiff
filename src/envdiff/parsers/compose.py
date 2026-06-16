from __future__ import annotations

import re
from pathlib import Path

from envdiff.models import EnvVarUsage, UsageScanResult
from envdiff.utils.ordering import sort_usages

INTERPOLATION_RE = re.compile(r"\$\{([A-Z0-9_]+)(?::-(.*?))?\}")


def scan_compose_file(path: str | Path) -> UsageScanResult:
    file_path = Path(path)
    usages: list[EnvVarUsage] = []

    for line_number, line in enumerate(file_path.read_text(encoding="utf-8").splitlines(), start=1):
        for match in INTERPOLATION_RE.finditer(line):
            default_value = match.group(2)
            usages.append(
                EnvVarUsage(
                    name=match.group(1),
                    file_path=str(file_path),
                    line_number=line_number,
                    usage_kind="compose_interpolation",
                    requiredness=(
                        "optional_with_default" if default_value is not None else "required"
                    ),
                    default_value=default_value,
                    source_type="compose",
                )
            )

    return UsageScanResult(usages=sort_usages(usages), warnings=())
