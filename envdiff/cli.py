from __future__ import annotations

from collections import Counter
from pathlib import Path

import typer

from envdiff.analyzers.baseline import (
    apply_suppressions,
    load_baseline_snapshot,
    load_ignore_keys,
    write_baseline_snapshot,
)
from envdiff.analyzers.compare import compare_dotenv_files
from envdiff.analyzers.doctor import doctor_repository
from envdiff.analyzers.scan import scan_repository
from envdiff.models import CommandMeta, JsonEnvelope, SummaryCounts
from envdiff.render.human import render_compare_result, render_doctor_result, render_scan_result
from envdiff.render.json import render_json

app = typer.Typer(
    add_completion=False,
    help="Analyze repository environment contracts deterministically.",
    no_args_is_help=True,
)


@app.command()
def compare(
    left: str = typer.Argument(..., help="Left dotenv file."),
    right: str = typer.Argument(..., help="Right dotenv file."),
    json_output: bool = typer.Option(False, "--json", help="Emit JSON output."),
) -> None:
    result = compare_dotenv_files(left, right)
    if json_output:
        envelope = JsonEnvelope(
            meta=CommandMeta(command="compare"),
            inputs={"left": left, "right": right},
            data=result,
        )
        typer.echo(render_json(envelope))
        return
    typer.echo(render_compare_result(result))


@app.command()
def scan(
    path: str = typer.Argument(..., help="Repository path to scan."),
    json_output: bool = typer.Option(False, "--json", help="Emit JSON output."),
) -> None:
    result = scan_repository(path)
    if json_output:
        envelope = JsonEnvelope(
            meta=CommandMeta(command="scan"),
            inputs={"path": path},
            data=result.model_dump(mode="json"),
        )
        typer.echo(render_json(envelope))
        return
    typer.echo(render_scan_result(result))


@app.command()
def doctor(
    path: str = typer.Argument(..., help="Repository path to validate."),
    fail_on: str = typer.Option("error", "--fail-on", help="Exit on severity threshold."),
    baseline: str | None = typer.Option(
        None,
        "--baseline",
        help="Suppress findings that match suppression keys in a baseline JSON file.",
    ),
    write_baseline: str | None = typer.Option(
        None,
        "--write-baseline",
        help="Write the current finding set to a baseline JSON file and exit successfully.",
    ),
    ignore_file: str | None = typer.Option(
        None,
        "--ignore-file",
        help="Path to a newline-delimited suppression file.",
    ),
    json_output: bool = typer.Option(False, "--json", help="Emit JSON output."),
) -> None:
    scan_result = scan_repository(path)
    findings = doctor_repository(scan_result)

    default_ignore_path = Path(path) / ".envdiffignore"
    suppression_keys = set()
    baseline_entry_count = 0

    if baseline:
        snapshot = load_baseline_snapshot(baseline)
        baseline_entry_count = len(snapshot.entries)
        suppression_keys.update(entry.suppression_key for entry in snapshot.entries)

    if ignore_file:
        suppression_keys.update(load_ignore_keys(ignore_file))
    elif default_ignore_path.is_file():
        ignore_file = str(default_ignore_path)
        suppression_keys.update(load_ignore_keys(default_ignore_path))

    active_findings, suppressed_findings = apply_suppressions(findings, suppression_keys)
    summary = _summarize(active_findings)

    baseline_written = None
    if write_baseline:
        write_baseline_snapshot(write_baseline, findings)
        baseline_written = write_baseline

    if json_output:
        envelope = JsonEnvelope(
            meta=CommandMeta(command="doctor"),
            inputs={
                "path": path,
                "fail_on": fail_on,
                "baseline": baseline,
                "write_baseline": write_baseline,
                "ignore_file": ignore_file,
            },
            summary=summary,
            findings=active_findings,
            data={
                "scan": scan_result.model_dump(mode="json"),
                "filtering": {
                    "baseline_entries": baseline_entry_count,
                    "suppressed_count": len(suppressed_findings),
                    "baseline_written": baseline_written,
                },
                "suppressed_findings": [
                    finding.model_dump(mode="json") for finding in suppressed_findings
                ],
            },
        )
        typer.echo(render_json(envelope))
    else:
        typer.echo(
            render_doctor_result(
                scan_result.root_path,
                active_findings,
                suppressed_count=len(suppressed_findings),
                baseline_written=baseline_written,
            )
        )

    if write_baseline:
        return

    if _should_fail(summary, fail_on):
        raise typer.Exit(code=2)


def main() -> None:
    app()


def _summarize(findings) -> SummaryCounts:
    counts = Counter(finding.severity for finding in findings)
    return SummaryCounts(
        error=counts.get("error", 0),
        warning=counts.get("warning", 0),
        info=counts.get("info", 0),
    )


def _should_fail(summary: SummaryCounts, threshold: str) -> bool:
    normalized = threshold.lower()
    if normalized == "error":
        return summary.error > 0
    if normalized == "warning":
        return summary.error > 0 or summary.warning > 0
    if normalized == "info":
        return summary.error > 0 or summary.warning > 0 or summary.info > 0
    raise typer.BadParameter("fail-on must be one of: error, warning, info")


if __name__ == "__main__":
    main()
