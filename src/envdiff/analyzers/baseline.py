from __future__ import annotations

import json
from pathlib import Path

from envdiff.models import BaselineEntry, BaselineSnapshot, Finding


def build_baseline_snapshot(findings: tuple[Finding, ...]) -> BaselineSnapshot:
    entries = []
    for finding in findings:
        if not finding.suppression_key:
            continue
        entries.append(
            BaselineEntry(
                suppression_key=finding.suppression_key,
                code=finding.code,
                severity=finding.severity,
                variable_name=finding.variable_name,
                title=finding.title,
                reason=finding.reason,
            )
        )

    ordered = tuple(sorted(entries, key=lambda entry: entry.suppression_key))
    return BaselineSnapshot(entries=ordered)


def write_baseline_snapshot(path: str | Path, findings: tuple[Finding, ...]) -> BaselineSnapshot:
    output_path = Path(path)
    snapshot = build_baseline_snapshot(findings)
    output_path.write_text(
        json.dumps(snapshot.model_dump(mode="json"), indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return snapshot


def load_baseline_snapshot(path: str | Path) -> BaselineSnapshot:
    payload = json.loads(Path(path).read_text(encoding="utf-8"))
    return BaselineSnapshot.model_validate(payload)


def load_ignore_keys(path: str | Path) -> set[str]:
    ignore_path = Path(path)
    keys = set()
    for raw_line in ignore_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        keys.add(line)
    return keys


def apply_suppressions(
    findings: tuple[Finding, ...],
    suppression_keys: set[str],
) -> tuple[tuple[Finding, ...], tuple[Finding, ...]]:
    active: list[Finding] = []
    suppressed: list[Finding] = []

    for finding in findings:
        if finding.suppression_key and finding.suppression_key in suppression_keys:
            suppressed.append(finding)
        else:
            active.append(finding)

    return tuple(active), tuple(suppressed)
