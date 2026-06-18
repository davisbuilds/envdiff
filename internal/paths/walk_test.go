package paths

import (
	"os"
	"path/filepath"
	"testing"
)

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

	found, err := FindNearestNamedFile(source, root, ".env")
	if err != nil {
		t.Fatalf("find nearest file: %v", err)
	}
	if found == nil || *found != marker {
		t.Fatalf("nearest = %v, want %s", found, marker)
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
