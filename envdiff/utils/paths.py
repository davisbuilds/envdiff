from __future__ import annotations

from collections.abc import Iterable
from pathlib import Path

DEFAULT_IGNORED_DIRS = {".git", ".venv", "__pycache__", "node_modules"}


def iter_repo_files(root: str | Path, ignore_dirs: Iterable[str] | None = None) -> tuple[Path, ...]:
    root_path = Path(root)
    ignored = set(ignore_dirs or DEFAULT_IGNORED_DIRS)
    files: list[Path] = []

    for path in root_path.rglob("*"):
        relative_parts = path.relative_to(root_path).parts
        if any(part in ignored for part in relative_parts):
            continue
        if path.is_file():
            files.append(path)

    return tuple(sorted(files))


def find_nearest_named_file(
    start_file: str | Path,
    root: str | Path,
    target_name: str,
) -> Path | None:
    current = Path(start_file).parent if Path(start_file).is_file() else Path(start_file)
    root_path = Path(root).resolve()

    while True:
        candidate = current / target_name
        if candidate.is_file():
            return candidate
        if current.resolve() == root_path:
            return None
        if current.parent == current:
            return None
        current = current.parent
