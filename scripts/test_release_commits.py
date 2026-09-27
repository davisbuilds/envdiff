"""Exercise syntax controls and actual Git history traversal."""

import os
from pathlib import Path
import subprocess
import sys
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
            valid = git("rev-parse", "HEAD")
            git("commit", "--allow-empty", "-qm", "Invalid retained change")
            git("checkout", "-qb", "main", base)
            git("merge", "--no-ff", "-qm", "Merge feature", "feature")
            previous = Path.cwd()
            try:
                os.chdir(directory)
                self.assertEqual(check_range(base, git("rev-parse", "HEAD")), ["Invalid retained change"])
                self.assertEqual(check_range(base, valid), [])
                self.assertEqual(check_range(base, base), [])
                for invalid in ["", "0" * 40, "HEAD", "g" * 40]:
                    with self.subTest(revision=invalid), self.assertRaises(ValueError):
                        check_range(invalid, valid)
                    with self.subTest(head=invalid), self.assertRaises(ValueError):
                        check_range(base, invalid)
                with self.assertRaises(subprocess.CalledProcessError):
                    check_range("f" * 40, valid)
                script = previous / "scripts" / "check_release_commits.py"
                env = dict(os.environ)
                env.pop("PR_TITLE", None)
                def run(*args):
                    return subprocess.run([sys.executable, str(script), *args], env=env, capture_output=True, text=True)
                self.assertEqual(run(base, valid, "--push").returncode, 0)
                self.assertNotEqual(run(base, git("rev-parse", "HEAD"), "--push").returncode, 0)
                self.assertNotEqual(run(base, valid, "--check-pr-title").returncode, 0)
                env["PR_TITLE"] = "feat: valid merge title"
                self.assertEqual(run(base, valid, "--check-pr-title").returncode, 0)
                self.assertNotEqual(run("0" * 40, valid, "--push").returncode, 0)
                self.assertNotEqual(run(base, base, "--push").returncode, 0)
                self.assertNotEqual(run(valid, base, "--push").returncode, 0)
                git("checkout", "-qb", "diverged", base)
                git("commit", "--allow-empty", "-qm", "fix: other branch")
                diverged = git("rev-parse", "HEAD")
                self.assertNotEqual(run(valid, diverged, "--push").returncode, 0)
                # A PR can diverge from an advanced base; its head-only commits
                # must still be validated without requiring push ancestry.
                self.assertEqual(run(valid, diverged, "--check-pr-title").returncode, 0)
            finally:
                os.chdir(previous)


if __name__ == "__main__":
    unittest.main()
