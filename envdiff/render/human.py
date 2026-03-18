from __future__ import annotations

from collections import Counter

from envdiff.models import Finding, RepoScanResult


def render_compare_result(result: dict[str, object]) -> str:
    lines = [
        f"Compare {result['left_path']} vs {result['right_path']}",
        f"Missing in left: {', '.join(result['missing_in_left']) or '-'}",
        f"Missing in right: {', '.join(result['missing_in_right']) or '-'}",
        f"Duplicates in left: {', '.join(result['duplicates_in_left']) or '-'}",
        f"Duplicates in right: {', '.join(result['duplicates_in_right']) or '-'}",
    ]

    differing = result["differing_values"]
    if differing:
        lines.append("Differing values:")
        for item in differing:
            lines.append(
                "  "
                f"{item['name']}: {item['left_kind']}={item['left_value']} vs "
                f"{item['right_kind']}={item['right_value']}"
            )

    return "\n".join(lines)


def render_scan_result(scan_result: RepoScanResult) -> str:
    lines = [
        f"Scan root: {scan_result.root_path}",
        f"Definitions: {len(scan_result.definitions)}",
        f"Usages: {len(scan_result.usages)}",
        f"Contracts: {len(scan_result.contracts)}",
    ]

    if scan_result.contracts:
        lines.append("Contracts:")
        for contract in scan_result.contracts:
            statuses = ",".join(contract.status) or "none"
            lines.append(f"  {contract.name} [{contract.requiredness}] ({statuses})")

    return "\n".join(lines)


def render_matrix_result(result: dict[str, object]) -> str:
    lines = [
        f"Matrix files: {result['file_count']}",
        f"Variables: {result['variable_count']}",
        f"Inconsistent: {result['inconsistent_variable_count']}",
    ]

    warnings = result["warnings"]
    if warnings:
        lines.append(f"Warnings: {len(warnings)}")

    variables = result["variables"]
    if not variables:
        lines.append("No variables matched the selected view.")
        return "\n".join(lines)

    lines.append("Variables:")
    for variable in variables:
        reasons = []
        if variable["missing_in"]:
            reasons.append("missing")
        if len(variable["value_kinds"]) > 1:
            reasons.append("kind-mismatch")
        if variable["duplicates_in"]:
            reasons.append("duplicate")
        lines.append(f"  {variable['name']} [{', '.join(reasons) or 'consistent'}]")
        for file_entry in variable["files"]:
            if file_entry["presence"] == "missing":
                lines.append(f"    {file_entry['path']}: missing")
                continue

            suffix = file_entry["value_kind"]
            if file_entry.get("is_duplicate"):
                suffix = f"{suffix}, duplicate"
            lines.append(f"    {file_entry['path']}: present ({suffix})")

    return "\n".join(lines)


def render_generate_result(
    variable_count: int,
    *,
    output_path: str | None = None,
    annotate: bool,
    check_path: str | None = None,
    check_matches: bool | None = None,
) -> str:
    if check_path is not None and check_matches is not None:
        if check_matches:
            return f"Generated output matches {check_path}"
        return f"Generated output drifted from {check_path}"

    suffix = " with annotations" if annotate else ""
    return (
        f"Generated {variable_count} variables{suffix} to {output_path}"
    )


def render_doctor_result(
    root_path: str,
    findings: tuple[Finding, ...],
    *,
    suppressed_count: int = 0,
    baseline_written: str | None = None,
) -> str:
    counts = Counter(finding.severity for finding in findings)
    lines = [
        f"Doctor root: {root_path}",
        f"Findings: {len(findings)}",
        (
            "Summary: "
            f"{counts.get('error', 0)} error, "
            f"{counts.get('warning', 0)} warning, "
            f"{counts.get('info', 0)} info"
        ),
    ]
    if suppressed_count:
        lines.append(f"Suppressed: {suppressed_count}")
    if baseline_written:
        lines.append(f"Baseline written: {baseline_written}")
    if not findings:
        lines.append("No active findings.")
        return "\n".join(lines)

    for severity in ("error", "warning", "info"):
        scoped = [finding for finding in findings if finding.severity == severity]
        if not scoped:
            continue
        lines.append(f"{severity.capitalize()}s:")
        for finding in scoped:
            lines.append(f"  {finding.code} {finding.details}")
    return "\n".join(lines)
