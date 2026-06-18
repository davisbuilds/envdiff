package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestRunPrintsHelpForNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	for _, command := range commandNames {
		if !strings.Contains(output, command) {
			t.Fatalf("help output missing command %q:\n%s", command, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsUnimplementedCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"doctor", "."}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "doctor is not implemented") {
		t.Fatalf("stderr = %q, want unimplemented message", stderr.String())
	}
}

func TestRunCompareJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{
			"compare",
			"tests/fixtures/compare/left.env",
			"tests/fixtures/compare/right.env",
			"--json",
		},
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	want := testutil.LoadGoldenJSON(t, "compare-basic.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunCompareHumanIncludesMissingHeading(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{
			"compare",
			"tests/fixtures/compare/left.env",
			"tests/fixtures/compare/right.env",
		},
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Missing in left") {
		t.Fatalf("stdout = %q, want Missing in left", stdout.String())
	}
}

func TestRunMatrixJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{
			"matrix",
			"tests/fixtures/matrix/a.env",
			"tests/fixtures/matrix/b.env",
			"tests/fixtures/matrix/c.env",
			"--json",
		},
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	want := testutil.LoadGoldenJSON(t, "matrix-basic.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunMatrixRejectsSinglePath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"matrix", "tests/fixtures/matrix/a.env"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "matrix requires at least two dotenv files") {
		t.Fatalf("stderr = %q, want matrix validation message", stderr.String())
	}
}

func TestRunScanJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"scan", "tests/fixtures/repos/simple_repo", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, "scan-simple-repo.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunGeneratePrintsDotenvToStdout(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"generate", "tests/fixtures/repos/simple_repo"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if stdout.String() != "DATABASE_URL=\nREDIS_URL=\nDEBUG=\n" {
		t.Fatalf("stdout = %q, want generated dotenv", stdout.String())
	}
}

func TestRunGenerateJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"generate", "tests/fixtures/repos/simple_repo", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, "generate-simple-repo.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunGenerateCheckDetectsDrift(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"generate", "tests/fixtures/repos/simple_repo", "--check"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "drifted from") {
		t.Fatalf("stdout = %q, want drift message", stdout.String())
	}
}

func TestRunGenerateCheckCanTargetExplicitFile(t *testing.T) {
	project := filepath.Join(t.TempDir(), "project")
	appDir := filepath.Join(project, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("create app dir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(project, ".env"),
		[]byte("DATABASE_URL=postgres://localhost/db\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, ".env.example"), []byte{}, 0o644); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(appDir, "main.py"),
		[]byte("import os\n\ndatabase_url = os.environ[\"DATABASE_URL\"]\n"),
		0o644,
	); err != nil {
		t.Fatalf("write app: %v", err)
	}
	targetPath := filepath.Join(t.TempDir(), "generated.env.example")
	if err := os.WriteFile(targetPath, []byte("DATABASE_URL=\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"generate", project, "--check", "--output", targetPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "matches") {
		t.Fatalf("stdout = %q, want matches", stdout.String())
	}
}
