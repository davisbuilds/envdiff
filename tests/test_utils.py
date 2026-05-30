from __future__ import annotations

from pathlib import Path

from src.utils.normalize import _is_int, is_placeholder, normalize_value_kind
from src.utils.paths import find_nearest_named_file, iter_repo_files


def test_normalize_value_kind_classifies_signed_numbers_and_placeholders() -> None:
    assert normalize_value_kind("-42") == "integer"
    assert normalize_value_kind("+3.14") == "float"
    assert is_placeholder("replace_me")
    assert not _is_int("")


def test_iter_repo_files_skips_default_and_custom_ignored_dirs(tmp_path: Path) -> None:
    visible = tmp_path / "app.py"
    visible.write_text("print('ok')\n")
    ignored = tmp_path / "ignored"
    ignored.mkdir()
    (ignored / ".env").write_text("SECRET=value\n")

    files = iter_repo_files(tmp_path, ignore_dirs=("ignored",))

    assert files == (visible,)


def test_find_nearest_named_file_searches_from_files_and_stops_at_root(tmp_path: Path) -> None:
    package = tmp_path / "package"
    nested = package / "src"
    nested.mkdir(parents=True)
    marker = package / ".env"
    marker.write_text("DATABASE_URL=postgres://db\n")
    source = nested / "settings.py"
    source.write_text("import os\n")

    assert find_nearest_named_file(source, tmp_path, ".env") == marker
    assert find_nearest_named_file(nested, nested, ".env") is None
    assert find_nearest_named_file(Path("/"), tmp_path, ".env") is None
