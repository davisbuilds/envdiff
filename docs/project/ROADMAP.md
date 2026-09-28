# Roadmap

envdiff remains a deterministic, repo-local environment-contract analyzer. The Go
CLI is the product; it does not load or distribute secrets or inspect machine-wide
shell configuration. [Architecture](../system/ARCHITECTURE.md),
[Features](../system/FEATURES.md), and the [JSON contract](../system/JSON_SCHEMA.md)
own current behavior.

## Current Direction

Keep CLI findings explainable and stable for team and CI use. Finding codes and the
JSON envelope are public contracts; expand detection from fixtures showing a useful
gap rather than broadening heuristics by default. The completed Python-to-Go port
left Go as the sole implementation and golden-output source. Application SemVer is
separate from JSON schema version `1`; [release operations](../system/RELEASES.md)
own versioning and recovery.

## Candidates, Not Scheduled Milestones

Parser expansion, editor diagnostics, and integration affordances are described
under [Features](../system/FEATURES.md#deferred-features). The
[Backlog](BACKLOG.md) holds concrete implementation friction and revisit conditions.
No order or release date is committed here.
