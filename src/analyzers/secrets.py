from __future__ import annotations

from pathlib import Path

from src.models import Finding, Location, RepoScanResult
from src.utils.normalize import is_non_empty_placeholder
from src.utils.ordering import sort_findings


def secret_and_placeholder_findings(scan_result: RepoScanResult) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    seen: set[tuple[str, str, int, str]] = set()

    for definition in scan_result.definitions:
        if Path(definition.file_path).name != ".env":
            continue

        if definition.normalized_value_kind == "secret_like":
            key = ("ENV008", definition.file_path, definition.line_number, definition.name)
            if key not in seen:
                seen.add(key)
                findings.append(
                    Finding(
                        code="ENV008",
                        severity="warning",
                        title="Secret-like committed value",
                        details=(
                            f"{definition.name} in {definition.file_path} looks like a real "
                            "secret value."
                        ),
                        variable_name=definition.name,
                        locations=(
                            Location(
                                file_path=definition.file_path,
                                line_number=definition.line_number,
                            ),
                        ),
                        confidence="low",
                        source_kind="heuristic",
                        reason="Value shape is long, opaque, and mixed-character.",
                        suppression_key=(
                            f"secret-like:{definition.file_path}:{definition.line_number}:{definition.name}"
                        ),
                    )
                )

        if is_non_empty_placeholder(definition.value):
            key = ("ENV009", definition.file_path, definition.line_number, definition.name)
            if key not in seen:
                seen.add(key)
                findings.append(
                    Finding(
                        code="ENV009",
                        severity="warning",
                        title="Placeholder-like committed value",
                        details=(
                            f"{definition.name} in {definition.file_path} uses a placeholder "
                            "value."
                        ),
                        variable_name=definition.name,
                        locations=(
                            Location(
                                file_path=definition.file_path,
                                line_number=definition.line_number,
                            ),
                        ),
                        confidence="low",
                        source_kind="heuristic",
                        reason=(
                            "Common placeholder values in committed .env files "
                            "are easy to miss."
                        ),
                        suppression_key=(
                            f"placeholder:{definition.file_path}:{definition.line_number}:{definition.name}"
                        ),
                    )
                )

    return sort_findings(findings)
