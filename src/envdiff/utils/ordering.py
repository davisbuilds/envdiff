from __future__ import annotations

from collections.abc import Iterable
from typing import TypeVar

from envdiff.models import EnvVarContract, EnvVarDefinition, EnvVarUsage, Finding

T = TypeVar("T")


def stable_sorted(items: Iterable[T], *, key: callable) -> tuple[T, ...]:
    return tuple(sorted(items, key=key))


def sort_definitions(definitions: Iterable[EnvVarDefinition]) -> tuple[EnvVarDefinition, ...]:
    return stable_sorted(
        definitions,
        key=lambda definition: (definition.name, definition.file_path, definition.line_number),
    )


def sort_usages(usages: Iterable[EnvVarUsage]) -> tuple[EnvVarUsage, ...]:
    return stable_sorted(
        usages,
        key=lambda usage: (
            usage.name,
            usage.file_path,
            usage.line_number if usage.line_number is not None else -1,
        ),
    )


def sort_contracts(contracts: Iterable[EnvVarContract]) -> tuple[EnvVarContract, ...]:
    return stable_sorted(contracts, key=lambda contract: contract.name)


def sort_findings(findings: Iterable[Finding]) -> tuple[Finding, ...]:
    severity_rank = {"error": 0, "warning": 1, "info": 2}
    return stable_sorted(
        findings,
        key=lambda finding: (
            severity_rank.get(finding.severity, 99),
            finding.code,
            finding.variable_name or "",
            tuple(
                (location.file_path, location.line_number or -1) for location in finding.locations
            ),
        ),
    )
