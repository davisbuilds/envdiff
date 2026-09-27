#!/usr/bin/env python3
"""Validate subjects retained by merge/rebase and the PR merge title."""

import argparse
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
    for revision in (base, head):
        if not re.fullmatch(r"[0-9a-f]{40}", revision) or revision == "0" * 40:
            raise ValueError("release validation requires nonzero full commit SHAs")
        subprocess.run(
            ["git", "cat-file", "-e", f"{revision}^{{commit}}"],
            check=True, capture_output=True,
        )
    # Actual merge commits carry no new consumer change; validate retained leaves.
    subjects = subprocess.check_output(
        ["git", "log", "--no-merges", "--format=%s", f"{base}..{head}"],
        text=True,
    ).splitlines()
    return [subject for subject in subjects if not valid_subject(subject)]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("base_sha")
    parser.add_argument("head_sha")
    parser.add_argument("--check-pr-title", action="store_true")
    args = parser.parse_args()
    try:
        invalid = check_range(args.base_sha, args.head_sha)
    except (ValueError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
    if args.check_pr_title:
        title = os.environ.get("PR_TITLE", "")
        if not valid_subject(title):
            invalid.append("PR title: " + title)
    if invalid:
        for subject in invalid:
            print(f"Invalid Conventional Commit subject: {subject}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
