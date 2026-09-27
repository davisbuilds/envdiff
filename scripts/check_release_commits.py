#!/usr/bin/env python3
"""Validate subjects retained by merge/rebase and the PR merge title."""

import os
import re
import subprocess
import sys

CONVENTIONAL = re.compile(
    r"^(feat|fix|perf|docs|test|chore|build|ci|style|refactor|revert)"
    r"(\([^\r\n()]+\))?!?: \S.*$"
)


def valid_subject(subject):
    return CONVENTIONAL.fullmatch(subject) is not None


def check_range(base, head):
    # Actual merge commits carry no new consumer change; validate retained leaves.
    subjects = subprocess.check_output(
        ["git", "log", "--no-merges", "--format=%s", f"{base}..{head}"],
        text=True,
    ).splitlines()
    return [subject for subject in subjects if not valid_subject(subject)]


def main():
    if len(sys.argv) != 3:
        sys.exit("usage: check_release_commits.py BASE_SHA HEAD_SHA")
    invalid = check_range(sys.argv[1], sys.argv[2])
    title = os.environ.get("PR_TITLE", "")
    if not valid_subject(title):
        invalid.append("PR title: " + title)
    if invalid:
        for subject in invalid:
            print(f"Invalid Conventional Commit subject: {subject}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
