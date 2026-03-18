from __future__ import annotations

from src.models import CommandMeta, Finding, JsonEnvelope, Location
from src.render.json import render_json


def test_json_envelope_has_schema_version() -> None:
    envelope = JsonEnvelope(meta=CommandMeta(command="scan"), inputs={"path": "."})

    assert envelope.schema_version() == "1"
    assert envelope.meta.schema_version == "1"


def test_json_rendering_is_stable_and_sorted() -> None:
    finding = Finding(
        code="ENV001",
        severity="warning",
        title="Example",
        details="Example finding",
        variable_name="DATABASE_URL",
        locations=(
            Location(file_path="b.py", line_number=3),
            Location(file_path="a.py", line_number=1),
        ),
    )
    envelope = JsonEnvelope(
        meta=CommandMeta(command="doctor"),
        inputs={"path": "."},
        findings=(finding,),
    )

    rendered = render_json(envelope)

    assert '"schema_version": "1"' in rendered
    assert rendered.index('"data"') < rendered.index('"findings"')
