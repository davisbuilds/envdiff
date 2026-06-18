package analyzers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeSyntheticRepo creates a repo of n Python files (each referencing a
// unique VAR/OTHER/THIRD triple) with the even-indexed VARs defined in a
// top-level .env, so doctor sees a mix of resolved and missing usages.
func writeSyntheticRepo(tb testing.TB, root string, n int) {
	tb.Helper()
	var env []byte
	for index := 0; index < n; index += 2 {
		env = append(env, fmt.Sprintf("VAR%d=value%d\n", index, index)...)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), env, 0o644); err != nil {
		tb.Fatalf("write .env: %v", err)
	}
	for index := 0; index < n; index++ {
		dir := filepath.Join(root, fmt.Sprintf("pkg%02d", index%50))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			tb.Fatalf("mkdir: %v", err)
		}
		content := fmt.Sprintf(
			"import os\n\nA = os.getenv(%q)\nB = os.environ[%q]\nC = os.getenv(%q, \"default\")\n",
			fmt.Sprintf("VAR%d", index),
			fmt.Sprintf("OTHER%d", index),
			fmt.Sprintf("THIRD%d", index),
		)
		path := filepath.Join(dir, fmt.Sprintf("mod%05d.py", index))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatalf("write py: %v", err)
		}
	}
}

func BenchmarkScanRepository(b *testing.B) {
	root := b.TempDir()
	writeSyntheticRepo(b, root, 300)

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := ScanRepository(root); err != nil {
			b.Fatalf("scan repository: %v", err)
		}
	}
}

func BenchmarkDoctorRepository(b *testing.B) {
	root := b.TempDir()
	writeSyntheticRepo(b, root, 2000)
	scanResult, err := ScanRepository(root)
	if err != nil {
		b.Fatalf("scan repository: %v", err)
	}

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		DoctorRepository(scanResult)
	}
}
