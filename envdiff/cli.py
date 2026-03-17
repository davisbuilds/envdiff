from __future__ import annotations

from collections import Counter

import typer

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
    json_output: bool = typer.Option(False, "--json", help="Emit JSON output."),
) -> None:
    scan_result = scan_repository(path)
    findings = doctor_repository(scan_result)
    summary = _summarize(findings)

    if json_output:
        envelope = JsonEnvelope(
            meta=CommandMeta(command="doctor"),
            inputs={"path": path, "fail_on": fail_on},
            summary=summary,
            findings=findings,
            data=scan_result.model_dump(mode="json"),
        )
        typer.echo(render_json(envelope))
    else:
        typer.echo(render_doctor_result(scan_result.root_path, findings))

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
