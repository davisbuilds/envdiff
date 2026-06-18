package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// RepoRoot returns the repository root from any Go test package.
func RepoRoot(t testing.TB) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if fileExists(filepath.Join(dir, "pyproject.toml")) && fileExists(filepath.Join(dir, "envdiff")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate envdiff repository root from %s", dir)
		}
		dir = parent
	}
}

// FixturePath returns an absolute path under tests/fixtures.
func FixturePath(t testing.TB, parts ...string) string {
	t.Helper()

	segments := append([]string{RepoRoot(t), "tests", "fixtures"}, parts...)
	return filepath.Join(segments...)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
