from __future__ import annotations

from envdiff.analyzers.aliases import canonical_name, find_alias_candidates


def test_canonical_name_expands_common_alias_forms() -> None:
    assert canonical_name("DB_URL") == "DATABASE_URL"
    assert canonical_name("PGHOST") == "POSTGRES_HOST"


def test_find_alias_candidates_detects_openai_key_drift() -> None:
    candidates = find_alias_candidates("OPENAI_API_KEY", {"OPENAI_KEY", "API_KEY"})

    assert candidates
    assert candidates[0].candidate_name == "OPENAI_KEY"

