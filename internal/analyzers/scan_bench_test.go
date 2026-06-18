package analyzers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScanRepository(b *testing.B) {
	root := b.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("VAR0=x\n"), 0o644); err != nil {
		b.Fatalf("write .env: %v", err)
	}

	const fileCount = 300
	for index := 0; index < fileCount; index++ {
		dir := filepath.Join(root, fmt.Sprintf("pkg%02d", index%20))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir: %v", err)
		}
		content := fmt.Sprintf(
			"import os\n\nA = os.getenv(%q)\nB = os.environ[%q]\nC = os.getenv(%q, \"default\")\n",
			fmt.Sprintf("VAR%d", index),
			fmt.Sprintf("OTHER%d", index),
			fmt.Sprintf("THIRD%d", index),
		)
		path := filepath.Join(dir, fmt.Sprintf("mod%03d.py", index))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatalf("write py: %v", err)
		}
	}

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := ScanRepository(root); err != nil {
			b.Fatalf("scan repository: %v", err)
		}
	}
}
