package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalResolvesSymlinkedComponents(t *testing.T) {
	real := t.TempDir()
	writeTestFile(t, filepath.Join(real, "marker"), "")
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	got, err := Canonical(link)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	want, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatalf("resolve real: %v", err)
	}
	if got != want {
		t.Fatalf("Canonical(%q) = %q, want %q", link, got, want)
	}
}

func TestCanonicalFallsBackForMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	got, err := Canonical(missing)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	if got != missing {
		t.Fatalf("Canonical(%q) = %q, want lexical fallback %q", missing, got, missing)
	}
}

func TestIterRepoFilesSkipsIgnoredDirsAndSortsResults(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "b.py"), "")
	writeTestFile(t, filepath.Join(root, "a.py"), "")
	writeTestFile(t, filepath.Join(root, "node_modules", "ignored.env"), "")
	writeTestFile(t, filepath.Join(root, ".git", "ignored.env"), "")

	files, err := IterRepoFiles(root, nil)
	if err != nil {
		t.Fatalf("iterate repo files: %v", err)
	}

	want := []string{filepath.Join(root, "a.py"), filepath.Join(root, "b.py")}
	if len(files) != len(want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
	for index := range want {
		if files[index] != want[index] {
			t.Fatalf("files = %v, want %v", files, want)
		}
	}
}

func TestFindNearestNamedFileSearchesFromFilesAndStopsAtRoot(t *testing.T) {
	root := t.TempDir()
	packageDir := filepath.Join(root, "package")
	nested := filepath.Join(packageDir, "src")
	source := filepath.Join(nested, "settings.py")
	marker := filepath.Join(packageDir, ".env")
	writeTestFile(t, source, "import os\n")
	writeTestFile(t, marker, "DATABASE_URL=postgres://db\n")

	// Paths are symlink-resolved (Canonical), so compare against the resolved
	// marker — on macOS $TMPDIR resolves /var -> /private/var.
	wantMarker, err := filepath.EvalSymlinks(marker)
	if err != nil {
		t.Fatalf("resolve marker: %v", err)
	}

	found, err := FindNearestNamedFile(source, root, ".env")
	if err != nil {
		t.Fatalf("find nearest file: %v", err)
	}
	if found == nil || *found != wantMarker {
		t.Fatalf("nearest = %v, want %s", found, wantMarker)
	}

	notFound, err := FindNearestNamedFile(nested, nested, ".env")
	if err != nil {
		t.Fatalf("find nearest file: %v", err)
	}
	if notFound != nil {
		t.Fatalf("nearest = %s, want nil", *notFound)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
