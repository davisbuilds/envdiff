from __future__ import annotations

import json

from src.analyzers.baseline import (
    apply_suppressions,
    build_baseline_snapshot,
    load_baseline_snapshot,
    load_ignore_keys,
    write_baseline_snapshot,
)
from src.analyzers.doctor import doctor_repository
from src.analyzers.scan import scan_repository


def test_baseline_snapshot_round_trip(tmp_path) -> None:
    findings = doctor_repository(scan_repository("tests/fixtures/doctor/project"))
    output_path = tmp_path / "baseline.json"

    written = write_baseline_snapshot(output_path, findings)
    loaded = load_baseline_snapshot(output_path)

    assert written == loaded
    payload = json.loads(output_path.read_text(encoding="utf-8"))
    assert payload["schema_version"] == "1"


def test_apply_suppressions_filters_by_suppression_key(tmp_path) -> None:
    findings = doctor_repository(scan_repository("tests/fixtures/doctor/project"))
    snapshot = build_baseline_snapshot(findings)
    target_key = snapshot.entries[0].suppression_key

    active, suppressed = apply_suppressions(findings, {target_key})

    assert any(finding.suppression_key == target_key for finding in suppressed)
    assert all(finding.suppression_key != target_key for finding in active)


def test_load_ignore_keys_skips_comments_and_blank_lines(tmp_path) -> None:
    ignore_file = tmp_path / ".envdiffignore"
    ignore_file.write_text(
        "# comment\n\nmissing:foo:BAR\nalias:OPENAI_API_KEY:OPENAI_KEY\n",
        encoding="utf-8",
    )

    keys = load_ignore_keys(ignore_file)

    assert keys == {"missing:foo:BAR", "alias:OPENAI_API_KEY:OPENAI_KEY"}
