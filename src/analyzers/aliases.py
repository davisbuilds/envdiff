from __future__ import annotations

from dataclasses import dataclass
from difflib import SequenceMatcher

TOKEN_EXPANSIONS = {
    "DB": ("DATABASE",),
    "PG": ("POSTGRES",),
}

POSTGRES_SUFFIXES = {"HOST", "PORT", "USER", "PASSWORD", "DATABASE", "DBNAME"}


@dataclass(frozen=True)
class AliasCandidate:
    candidate_name: str
    score: float
    reason: str


def find_alias_candidates(
    missing_name: str,
    defined_names: set[str],
    *,
    threshold: float = 0.8,
) -> tuple[AliasCandidate, ...]:
    candidates: list[AliasCandidate] = []
    missing_canonical = canonical_name(missing_name)
    missing_tokens = canonical_tokens(missing_name)

    for defined_name in sorted(defined_names):
        if defined_name == missing_name:
            continue

        defined_canonical = canonical_name(defined_name)
        defined_tokens = canonical_tokens(defined_name)
        token_overlap = _jaccard(missing_tokens, defined_tokens)
        sequence_ratio = SequenceMatcher(None, missing_canonical, defined_canonical).ratio()

        if missing_canonical == defined_canonical:
            candidates.append(
                AliasCandidate(
                    candidate_name=defined_name,
                    score=0.99,
                    reason=(
                        f"Canonical token expansion matches: {missing_canonical} == "
                        f"{defined_canonical}."
                    ),
                )
            )
            continue

        if token_overlap >= (2 / 3) and sequence_ratio >= threshold:
            candidates.append(
                AliasCandidate(
                    candidate_name=defined_name,
                    score=round(sequence_ratio, 2),
                    reason=(
                        f"Token overlap {token_overlap:.2f} and name similarity "
                        f"{sequence_ratio:.2f} suggest drift."
                    ),
                )
            )

    return tuple(sorted(candidates, key=lambda item: (-item.score, item.candidate_name)))


def canonical_name(name: str) -> str:
    return "_".join(canonical_tokens(name))


def canonical_tokens(name: str) -> tuple[str, ...]:
    upper_name = name.upper()
    if upper_name.startswith("PG") and upper_name[2:] in POSTGRES_SUFFIXES:
        suffix = upper_name[2:]
        return ("POSTGRES", "DATABASE" if suffix == "DBNAME" else suffix)

    raw_tokens = upper_name.split("_")
    tokens: list[str] = []
    for token in raw_tokens:
        if token in TOKEN_EXPANSIONS:
            tokens.extend(TOKEN_EXPANSIONS[token])
        else:
            tokens.append(token)
    return tuple(tokens)


def _jaccard(left: tuple[str, ...], right: tuple[str, ...]) -> float:
    left_set = set(left)
    right_set = set(right)
    union = left_set | right_set
    if not union:
        return 0.0
    return len(left_set & right_set) / len(union)

