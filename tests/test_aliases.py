from __future__ import annotations

from envdiff.analyzers.aliases import _jaccard, canonical_name, find_alias_candidates


def test_canonical_name_expands_common_alias_forms() -> None:
    assert canonical_name("DB_URL") == "DATABASE_URL"
    assert canonical_name("PGHOST") == "POSTGRES_HOST"


def test_find_alias_candidates_detects_openai_key_drift() -> None:
    candidates = find_alias_candidates("OPENAI_API_KEY", {"OPENAI_KEY", "API_KEY"})

    assert candidates
    assert candidates[0].candidate_name == "OPENAI_KEY"


def test_find_alias_candidates_skips_exact_names_and_reports_canonical_matches() -> None:
    candidates = find_alias_candidates("DATABASE_URL", {"DATABASE_URL", "DB_URL"})

    assert len(candidates) == 1
    assert candidates[0].candidate_name == "DB_URL"
    assert candidates[0].score == 0.99
    assert "Canonical token expansion matches" in candidates[0].reason


def test_jaccard_returns_zero_for_empty_token_sets() -> None:
    assert _jaccard((), ()) == 0.0
