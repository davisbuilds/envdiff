# Contributing

Bug reports, focused fixes, documentation improvements, and supported proposals
are welcome. Discuss substantial parser expansion, new dependencies, or changes to
the CLI and JSON/finding contracts before investing in implementation. This is a
solo-maintained project; a proposal does not imply a support or response-time promise.

Agent-assisted work is welcome. Submitters should understand the change's intent,
important behavior, tradeoffs, and verification, and explain any limits in the PR.
No prompt transcript or manual rewrite is required. A clear [Backlog](docs/project/BACKLOG.md)
entry can go directly to a PR; use an issue when discussion or coordination helps.

Work on a focused branch from `main`. Describe the problem and resulting behavior
with relevant test evidence. [Operations](docs/system/OPERATIONS.md) has setup and
checks; [Git policy](docs/project/GIT_HISTORY_POLICY.md) owns retained Conventional
Commits and merge strategy. Use `!` or a `BREAKING CHANGE:` footer and explain
migration for incompatible behavior. Release Please drafts `CHANGELOG.md`; review
consumer meaning and migration steps in its release PR rather than editing a
parallel release log for ordinary PRs.
