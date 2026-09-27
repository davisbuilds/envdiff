"""Exercise syntax controls and actual Git history traversal."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest

from check_release_commits import check_range, valid_subject


class ReleaseCommitsTest(unittest.TestCase):
    def test_subjects(self):
        for subject in ["feat: add version", "fix(cli)!: change output", "chore(main): release 0.1.0", "revert: remove change"]:
            with self.subTest(subject=subject):
                self.assertTrue(valid_subject(subject))
        for subject in ["Add version", "fix:", "fix: ", "feat(): no scope", "fix: ok\nmalicious", "fixup! fix: change"]:
            with self.subTest(subject=subject):
                self.assertFalse(valid_subject(subject))

    def test_retained_history_excludes_merges(self):
        with tempfile.TemporaryDirectory() as directory:
            def git(*args):
                return subprocess.check_output(["git", "-C", directory, *args], text=True).strip()

            git("init", "-q", "--initial-branch=base")
            git("config", "user.email", "test@example.com")
            git("config", "user.name", "Test")
            git("config", "commit.gpgsign", "false")
            git("commit", "--allow-empty", "-qm", "Historic nonconventional subject")
            base = git("rev-parse", "HEAD")
            git("checkout", "-qb", "feature")
            git("commit", "--allow-empty", "-qm", "feat: add version")
            git("commit", "--allow-empty", "-qm", "Invalid retained change")
            git("checkout", "-qb", "main", base)
            git("merge", "--no-ff", "-qm", "Merge feature", "feature")
            previous = Path.cwd()
            try:
                os.chdir(directory)
                self.assertEqual(check_range(base, "HEAD"), ["Invalid retained change"])
                self.assertEqual(check_range(base, base), [])
            finally:
                os.chdir(previous)


if __name__ == "__main__":
    unittest.main()
