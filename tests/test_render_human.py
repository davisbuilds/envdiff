from __future__ import annotations

from src.models import EnvVarContract, EnvVarDefinition, EnvVarUsage, Finding, RepoScanResult
from src.render.human import (
    render_compare_result,
    render_doctor_result,
    render_generate_result,
    render_matrix_result,
    render_scan_result,
)


def test_compare_result_renders_missing_duplicates_and_differing_values() -> None:
    rendered = render_compare_result(
        {
            "left_path": ".env",
            "right_path": ".env.example",
            "missing_in_left": ["API_KEY"],
            "missing_in_right": [],
            "duplicates_in_left": ["PORT"],
            "duplicates_in_right": [],
            "differing_values": [
                {
                    "name": "DEBUG",
                    "left_kind": "literal",
                    "left_value": "true",
                    "right_kind": "placeholder",
                    "right_value": "<debug>",
                }
            ],
        }
    )

    assert rendered == "\n".join(
        [
            "Compare .env vs .env.example",
            "Missing in left: API_KEY",
            "Missing in right: -",
            "Duplicates in left: PORT",
            "Duplicates in right: -",
            "Differing values:",
            "  DEBUG: literal=true vs placeholder=<debug>",
        ]
    )


def test_scan_result_renders_contract_summary_with_statuses() -> None:
    rendered = render_scan_result(
        RepoScanResult(
            root_path="/repo",
            definitions=(
                EnvVarDefinition(
                    name="DATABASE_URL",
                    value="postgres://db",
                    normalized_value_kind="literal",
                    file_path=".env",
                    line_number=1,
                    source_type="dotenv",
                ),
            ),
            usages=(
                EnvVarUsage(
                    name="DATABASE_URL",
                    file_path="app.py",
                    line_number=2,
                    usage_kind="read",
                    requiredness="required",
                    source_type="python",
                ),
            ),
            contracts=(
                EnvVarContract(
                    name="DATABASE_URL",
                    requiredness="required",
                    status=("defined", "used"),
                ),
                EnvVarContract(name="OPTIONAL_FLAG", requiredness="optional"),
            ),
        )
    )

    assert rendered == "\n".join(
        [
            "Scan root: /repo",
            "Definitions: 1",
            "Usages: 1",
            "Contracts: 2",
            "Contracts:",
            "  DATABASE_URL [required] (defined,used)",
            "  OPTIONAL_FLAG [optional] (none)",
        ]
    )


def test_matrix_result_renders_empty_view_with_warning_count() -> None:
    rendered = render_matrix_result(
        {
            "file_count": 2,
            "variable_count": 0,
            "inconsistent_variable_count": 0,
            "warnings": ["ignored file"],
            "variables": [],
        }
    )

    assert rendered == "\n".join(
        [
            "Matrix files: 2",
            "Variables: 0",
            "Inconsistent: 0",
            "Warnings: 1",
            "No variables matched the selected view.",
        ]
    )


def test_matrix_result_renders_variable_reasons_and_file_presence() -> None:
    rendered = render_matrix_result(
        {
            "file_count": 2,
            "variable_count": 2,
            "inconsistent_variable_count": 1,
            "warnings": [],
            "variables": [
                {
                    "name": "DATABASE_URL",
                    "missing_in": [".env.example"],
                    "value_kinds": ["literal", "placeholder"],
                    "duplicates_in": [".env"],
                    "files": [
                        {
                            "path": ".env",
                            "presence": "present",
                            "value_kind": "literal",
                            "is_duplicate": True,
                        },
                        {
                            "path": ".env.example",
                            "presence": "missing",
                        },
                    ],
                },
                {
                    "name": "LOG_LEVEL",
                    "missing_in": [],
                    "value_kinds": ["literal"],
                    "duplicates_in": [],
                    "files": [
                        {
                            "path": ".env",
                            "presence": "present",
                            "value_kind": "literal",
                        },
                    ],
                },
            ],
        }
    )

    assert rendered == "\n".join(
        [
            "Matrix files: 2",
            "Variables: 2",
            "Inconsistent: 1",
            "Variables:",
            "  DATABASE_URL [missing, kind-mismatch, duplicate]",
            "    .env: present (literal, duplicate)",
            "    .env.example: missing",
            "  LOG_LEVEL [consistent]",
            "    .env: present (literal)",
        ]
    )


def test_generate_result_renders_check_and_write_modes() -> None:
    assert (
        render_generate_result(3, annotate=False, check_path=".env.example", check_matches=True)
        == "Generated output matches .env.example"
    )
    assert (
        render_generate_result(3, annotate=False, check_path=".env.example", check_matches=False)
        == "Generated output drifted from .env.example"
    )
    assert (
        render_generate_result(3, output_path=".env.example", annotate=True)
        == "Generated 3 variables with annotations to .env.example"
    )


def test_doctor_result_renders_empty_findings_with_suppression_metadata() -> None:
    rendered = render_doctor_result(
        "/repo",
        (),
        suppressed_count=2,
        baseline_written=".envdiff-baseline.json",
    )

    assert rendered == "\n".join(
        [
            "Doctor root: /repo",
            "Findings: 0",
            "Summary: 0 error, 0 warning, 0 info",
            "Suppressed: 2",
            "Baseline written: .envdiff-baseline.json",
            "No active findings.",
        ]
    )


def test_doctor_result_groups_findings_by_severity() -> None:
    rendered = render_doctor_result(
        "/repo",
        (
            Finding(
                code="ENV001",
                severity="warning",
                title="Missing definition",
                details="DATABASE_URL is used but not defined",
            ),
            Finding(
                code="ENV002",
                severity="error",
                title="Duplicate definition",
                details="PORT is defined twice",
            ),
        ),
    )

    assert rendered == "\n".join(
        [
            "Doctor root: /repo",
            "Findings: 2",
            "Summary: 1 error, 1 warning, 0 info",
            "Errors:",
            "  ENV002 PORT is defined twice",
            "Warnings:",
            "  ENV001 DATABASE_URL is used but not defined",
        ]
    )
