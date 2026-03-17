from __future__ import annotations

from typing import Any

from pydantic import BaseModel, ConfigDict, Field

SCHEMA_VERSION = "1"


class ModelBase(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)


class CommandMeta(ModelBase):
    command: str
    schema_version: str = Field(default=SCHEMA_VERSION)


class Location(ModelBase):
    file_path: str
    line_number: int | None = None
    column_number: int | None = None


class EnvVarDefinition(ModelBase):
    name: str
    value: str
    normalized_value_kind: str
    file_path: str
    line_number: int
    source_type: str
    is_duplicate: bool = False
    parse_warnings: tuple[str, ...] = ()


class EnvVarUsage(ModelBase):
    name: str
    file_path: str
    line_number: int | None = None
    usage_kind: str
    requiredness: str
    default_value: str | None = None
    source_type: str


class EnvVarContract(ModelBase):
    name: str
    definitions: tuple[EnvVarDefinition, ...] = ()
    usages: tuple[EnvVarUsage, ...] = ()
    requiredness: str = "unknown"
    aliases: tuple[str, ...] = ()
    secret_like: bool = False
    placeholder_like: bool = False
    status: tuple[str, ...] = ()
    resolution_notes: tuple[str, ...] = ()


class DotenvParseResult(ModelBase):
    definitions: tuple[EnvVarDefinition, ...] = ()
    warnings: tuple[str, ...] = ()


class UsageScanResult(ModelBase):
    usages: tuple[EnvVarUsage, ...] = ()
    warnings: tuple[str, ...] = ()


class ResolutionDecision(ModelBase):
    source_file: str
    env_file: str | None = None
    example_file: str | None = None
    notes: tuple[str, ...] = ()


class RepoScanResult(ModelBase):
    root_path: str
    definitions: tuple[EnvVarDefinition, ...] = ()
    usages: tuple[EnvVarUsage, ...] = ()
    contracts: tuple[EnvVarContract, ...] = ()
    resolutions: tuple[ResolutionDecision, ...] = ()
    warnings: tuple[str, ...] = ()


class Finding(ModelBase):
    code: str
    severity: str
    title: str
    details: str
    variable_name: str | None = None
    locations: tuple[Location, ...] = ()
    related_variables: tuple[str, ...] = ()
    suggested_fix: str | None = None
    confidence: str | None = None
    source_kind: str = "deterministic"
    reason: str | None = None
    suppression_key: str | None = None


class SummaryCounts(ModelBase):
    error: int = 0
    warning: int = 0
    info: int = 0


class JsonEnvelope(ModelBase):
    meta: CommandMeta
    inputs: dict[str, Any]
    summary: SummaryCounts = Field(default_factory=SummaryCounts)
    findings: tuple[Finding, ...] = ()
    data: dict[str, Any] = Field(default_factory=dict)

    @classmethod
    def schema_version(cls) -> str:
        return SCHEMA_VERSION
