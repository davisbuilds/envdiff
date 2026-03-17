from __future__ import annotations

from urllib.parse import urlparse

PLACEHOLDER_VALUES = {
    "",
    "changeme",
    "your_key_here",
    "example",
    "example_value",
    "replace_me",
}


def normalize_value_kind(value: str) -> str:
    stripped = value.strip()
    lowered = stripped.lower()

    if lowered in PLACEHOLDER_VALUES:
        return "placeholder"

    if lowered in {"true", "false"}:
        return "boolean"

    if _is_int(stripped):
        return "integer"

    if _is_float(stripped):
        return "float"

    if _looks_like_url(stripped):
        return "url"

    if _looks_like_secret(stripped):
        return "secret_like"

    return "string"


def is_placeholder(value: str) -> bool:
    return normalize_value_kind(value) == "placeholder"


def is_non_empty_placeholder(value: str) -> bool:
    return value.strip() != "" and normalize_value_kind(value) == "placeholder"


def _is_int(value: str) -> bool:
    if not value:
        return False
    if value.startswith(("+", "-")):
        value = value[1:]
    return value.isdigit()


def _is_float(value: str) -> bool:
    if not value or value.count(".") != 1:
        return False
    left, right = value.split(".", 1)
    if left.startswith(("+", "-")):
        left = left[1:]
    return bool(left) and bool(right) and left.isdigit() and right.isdigit()


def _looks_like_url(value: str) -> bool:
    if "://" not in value:
        return False
    parsed = urlparse(value)
    return bool(parsed.scheme and parsed.netloc)


def _looks_like_secret(value: str) -> bool:
    if len(value) < 20:
        return False
    alpha = sum(character.isalpha() for character in value)
    digits = sum(character.isdigit() for character in value)
    return alpha > 0 and digits > 0
