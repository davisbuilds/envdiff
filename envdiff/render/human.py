from __future__ import annotations

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


def render_doctor_result(root_path: str, findings: tuple[Finding, ...]) -> str:
    lines = [f"Doctor root: {root_path}", f"Findings: {len(findings)}"]
    for finding in findings:
        lines.append(f"{finding.severity.upper()} {finding.code} {finding.details}")
    return "\n".join(lines)
