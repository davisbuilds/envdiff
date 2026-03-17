from __future__ import annotations

import json

from envdiff.models import JsonEnvelope


def render_json(envelope: JsonEnvelope) -> str:
    return json.dumps(envelope.model_dump(mode="json"), indent=2, sort_keys=True)

