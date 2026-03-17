from __future__ import annotations

import ast
from pathlib import Path

from envdiff.models import EnvVarUsage, UsageScanResult
from envdiff.utils.ordering import sort_usages


class _PythonEnvVisitor(ast.NodeVisitor):
    def __init__(self, file_path: str) -> None:
        self.file_path = file_path
        self.usages: list[EnvVarUsage] = []

    def visit_Subscript(self, node: ast.Subscript) -> None:
        if _is_os_environ(node.value):
            name = _string_value(node.slice)
            if name:
                self.usages.append(
                    EnvVarUsage(
                        name=name,
                        file_path=self.file_path,
                        line_number=getattr(node, "lineno", None),
                        usage_kind="os.environ",
                        requiredness="required",
                        source_type="python",
                    )
                )
        self.generic_visit(node)

    def visit_Call(self, node: ast.Call) -> None:
        if _is_os_getenv(node.func) and node.args:
            name = _string_value(node.args[0])
            if name:
                default_value = _string_value(node.args[1]) if len(node.args) > 1 else None
                self.usages.append(
                    EnvVarUsage(
                        name=name,
                        file_path=self.file_path,
                        line_number=getattr(node, "lineno", None),
                        usage_kind="os.getenv",
                        requiredness=(
                            "optional_with_default" if default_value is not None else "optional"
                        ),
                        default_value=default_value,
                        source_type="python",
                    )
                )
        self.generic_visit(node)


def scan_python_file(path: str | Path) -> UsageScanResult:
    file_path = Path(path)
    tree = ast.parse(file_path.read_text(encoding="utf-8"), filename=str(file_path))
    visitor = _PythonEnvVisitor(str(file_path))
    visitor.visit(tree)
    return UsageScanResult(usages=sort_usages(visitor.usages), warnings=())


def _is_os_environ(node: ast.AST) -> bool:
    return (
        isinstance(node, ast.Attribute)
        and isinstance(node.value, ast.Name)
        and node.value.id == "os"
        and node.attr == "environ"
    )


def _is_os_getenv(node: ast.AST) -> bool:
    return (
        isinstance(node, ast.Attribute)
        and isinstance(node.value, ast.Name)
        and node.value.id == "os"
        and node.attr == "getenv"
    )


def _string_value(node: ast.AST | None) -> str | None:
    if isinstance(node, ast.Constant) and isinstance(node.value, str):
        return node.value
    return None
