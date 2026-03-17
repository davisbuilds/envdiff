from __future__ import annotations

from collections import defaultdict

from envdiff.analyzers.aliases import find_alias_candidates
from envdiff.analyzers.secrets import secret_and_placeholder_findings
from envdiff.models import Finding, Location, RepoScanResult, ResolutionDecision
from envdiff.utils.ordering import sort_findings


def doctor_repository(scan_result: RepoScanResult) -> tuple[Finding, ...]:
    definitions_by_file = defaultdict(list)
    resolutions_by_source = {
        resolution.source_file: resolution for resolution in scan_result.resolutions
    }
    associated_usage_names = defaultdict(set)
    findings: list[Finding] = []
    seen: set[tuple[str, str | None, str]] = set()

    for definition in scan_result.definitions:
        definitions_by_file[definition.file_path].append(definition)
        if definition.is_duplicate:
            findings.append(
                Finding(
                    code="ENV006",
                    severity="warning",
                    title="Duplicate definition",
                    details=(
                        f"{definition.name} is defined more than once in "
                        f"{definition.file_path}."
                    ),
                    variable_name=definition.name,
                    locations=(
                        Location(
                            file_path=definition.file_path,
                            line_number=definition.line_number,
                        ),
                    ),
                    reason="Duplicate keys create ambiguous effective values.",
                    suppression_key=f"duplicate:{definition.file_path}:{definition.name}",
                )
            )

    for usage in scan_result.usages:
        resolution = resolutions_by_source.get(usage.file_path)
        env_names = _definition_names(
            definitions_by_file, resolution.env_file if resolution else None
        )
        example_names = _definition_names(
            definitions_by_file, resolution.example_file if resolution else None
        )

        if resolution:
            if resolution.env_file:
                associated_usage_names[resolution.env_file].add(usage.name)
            if resolution.example_file:
                associated_usage_names[resolution.example_file].add(usage.name)

        if usage.name not in env_names:
            code = "ENV001" if usage.requiredness == "required" else "ENV002"
            severity = "error" if usage.requiredness == "required" else "warning"
            key = (code, usage.name, usage.file_path)
            if key not in seen:
                seen.add(key)
                findings.append(
                    Finding(
                        code=code,
                        severity=severity,
                        title="Missing variable",
                        details=_missing_details(usage, resolution),
                        variable_name=usage.name,
                        locations=(
                            Location(file_path=usage.file_path, line_number=usage.line_number),
                        ),
                        reason="The associated .env file does not define a referenced variable.",
                        suppression_key=f"missing:{usage.file_path}:{usage.name}",
                    )
                )
            findings.extend(
                _alias_findings_for_missing_usage(
                    usage_name=usage.name,
                    usage_path=usage.file_path,
                    usage_line=usage.line_number,
                    env_names=env_names,
                    seen=seen,
                )
            )

        if resolution and resolution.example_file and usage.name not in example_names:
            key = ("ENV004", usage.name, resolution.example_file)
            if key not in seen:
                seen.add(key)
                findings.append(
                    Finding(
                        code="ENV004",
                        severity="warning",
                        title="Undocumented variable",
                        details=(
                            f"{usage.name} is referenced by {usage.file_path} but absent from "
                            f"{resolution.example_file}."
                        ),
                        variable_name=usage.name,
                        locations=(
                            Location(file_path=usage.file_path, line_number=usage.line_number),
                        ),
                        reason="Referenced variables should appear in the nearest .env.example.",
                        suppression_key=f"undocumented:{resolution.example_file}:{usage.name}",
                    )
                )

    for definition in scan_result.definitions:
        associated_names = associated_usage_names[definition.file_path]
        if definition.name not in associated_names:
            key = ("ENV003", definition.name, definition.file_path)
            if key in seen:
                continue
            seen.add(key)
            findings.append(
                Finding(
                    code="ENV003",
                    severity="info",
                    title="Unused variable",
                    details=(
                        f"{definition.name} is defined in {definition.file_path} "
                        "but not referenced."
                    ),
                    variable_name=definition.name,
                    locations=(
                        Location(
                            file_path=definition.file_path,
                            line_number=definition.line_number,
                        ),
                    ),
                    reason="Defined variables without matching usage may be stale.",
                    suppression_key=f"unused:{definition.file_path}:{definition.name}",
                )
            )

    findings.extend(_skew_findings(scan_result, definitions_by_file, seen))
    findings.extend(secret_and_placeholder_findings(scan_result))
    return sort_findings(findings)


def _definition_names(definitions_by_file, file_path: str | None) -> set[str]:
    if not file_path:
        return set()
    return {definition.name for definition in definitions_by_file[file_path]}


def _missing_details(usage, resolution: ResolutionDecision | None) -> str:
    if resolution and resolution.env_file:
        return (
            f"{usage.name} is referenced by {usage.file_path} but missing from "
            f"{resolution.env_file}."
        )
    return f"{usage.name} is referenced by {usage.file_path} but no associated .env defines it."


def _skew_findings(scan_result: RepoScanResult, definitions_by_file, seen) -> list[Finding]:
    findings: list[Finding] = []
    for resolution in scan_result.resolutions:
        if not resolution.env_file or not resolution.example_file:
            continue
        env_names = _definition_names(definitions_by_file, resolution.env_file)
        example_names = _definition_names(definitions_by_file, resolution.example_file)

        for name in sorted(env_names - example_names):
            key = ("ENV005", name, resolution.env_file)
            if key in seen:
                continue
            seen.add(key)
            findings.append(
                Finding(
                    code="ENV005",
                    severity="warning",
                    title="Environment skew",
                    details=(
                        f"{name} is present in {resolution.env_file} but absent from "
                        f"{resolution.example_file}."
                    ),
                    variable_name=name,
                    locations=(Location(file_path=resolution.env_file),),
                    reason=(
                        "Nearest .env and .env.example should describe the same "
                        "contract surface."
                    ),
                    suppression_key=f"skew:{resolution.env_file}:{name}",
                )
            )
    return findings


def _alias_findings_for_missing_usage(
    *,
    usage_name: str,
    usage_path: str,
    usage_line: int | None,
    env_names: set[str],
    seen,
) -> list[Finding]:
    findings: list[Finding] = []
    for candidate in find_alias_candidates(usage_name, env_names):
        key = ("ENV007", usage_name, candidate.candidate_name)
        if key in seen:
            continue
        seen.add(key)
        findings.append(
            Finding(
                code="ENV007",
                severity="warning",
                title="Possible alias candidate",
                details=(
                    f"{usage_name} is missing, but nearby definitions include "
                    f"{candidate.candidate_name}."
                ),
                variable_name=usage_name,
                locations=(Location(file_path=usage_path, line_number=usage_line),),
                related_variables=(candidate.candidate_name,),
                confidence="low",
                source_kind="heuristic",
                reason=candidate.reason,
                suppression_key=f"alias:{usage_name}:{candidate.candidate_name}",
            )
        )
    return findings
