from __future__ import annotations

import re
from pathlib import Path

from envdiff.models import EnvVarUsage, UsageScanResult
from envdiff.utils.ordering import sort_usages

EXPRESSION_RE = re.compile(r"\$\{\{\s*(.*?)\s*\}\}")
ACTIONS_REFERENCE_RE = re.compile(r"\b(secrets|vars)\.([A-Z_][A-Z0-9_]*)\b")


def scan_github_actions_file(path: str | Path) -> UsageScanResult:
    file_path = Path(path)
    usages: list[EnvVarUsage] = []

    for line_number, line in enumerate(file_path.read_text(encoding="utf-8").splitlines(), start=1):
        for expression in EXPRESSION_RE.findall(line):
            default_value = _extract_default(expression)
            requiredness = "optional_with_default" if default_value is not None else "required"
            for reference_type, name in ACTIONS_REFERENCE_RE.findall(expression):
                usages.append(
                    EnvVarUsage(
                        name=name,
                        file_path=str(file_path),
                        line_number=line_number,
                        usage_kind=f"github_actions_{reference_type}",
                        requiredness=requiredness,
                        default_value=default_value,
                        source_type="github_actions",
                    )
                )

    return UsageScanResult(usages=sort_usages(usages), warnings=())


def _extract_default(expression: str) -> str | None:
    if "||" not in expression:
        return None

    fallback = expression.split("||", 1)[1].strip()
    if not fallback:
        return None
    if len(fallback) >= 2 and fallback[0] == fallback[-1] and fallback[0] in {'"', "'"}:
        return fallback[1:-1]
    return fallback
