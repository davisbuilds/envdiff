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

func TestRunPrintsHelpForExplicitHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	for _, want := range []string{"compare", "generate", "matrix", "scan", "doctor"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output missing %q:\n%s", want, output)
		}
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"bogus"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
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

func TestRunMatrixShowAllJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{
			"matrix",
			"tests/fixtures/matrix/a.env",
			"tests/fixtures/matrix/b.env",
			"tests/fixtures/matrix/c.env",
			"--show-all",
			"--json",
		},
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	want := testutil.LoadGoldenJSON(t, "matrix-show-all.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunMatrixRejectsSinglePath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"matrix", "tests/fixtures/matrix/a.env"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (usage error)", code)
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

func TestRunScanHumanIncludesContractCount(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"scan", "tests/fixtures/repos/simple_repo"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Contracts: 3") {
		t.Fatalf("stdout = %q, want Contracts: 3", stdout.String())
	}
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

func TestRunGenerateAnnotatedOutputWritesFile(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	outputPath := filepath.Join(t.TempDir(), ".env.example")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{
			"generate",
			"tests/fixtures/repos/simple_repo",
			"--annotate",
			"--output",
			outputPath,
		},
		&stdout,
		&stderr,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(content), "# Generated by envdiff. Review before committing.") {
		t.Fatalf("output content = %q", string(content))
	}
	if !strings.Contains(stdout.String(), "Generated 3 variables with annotations") {
		t.Fatalf("stdout = %q", stdout.String())
	}
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

func TestRunGenerateCheckJSONExitsTwoOnDrift(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{"generate", "tests/fixtures/repos/simple_repo", "--check", "--json"},
		&stdout,
		&stderr,
	)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"matches\": false") {
		t.Fatalf("stdout = %q, want JSON check payload", stdout.String())
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

func TestRunDoctorJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"doctor", "tests/fixtures/doctor/project", "--json"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, "doctor-project.json")
	testutil.AssertJSONEqual(t, got, want)
}

func TestRunDoctorThresholdAndInvalidFailOn(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{"doctor", "tests/fixtures/doctor/project", "--fail-on", "warning"},
		&stdout,
		&stderr,
	)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ENV001") {
		t.Fatalf("stdout = %q, want ENV001", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"doctor", "tests/fixtures/doctor/project", "--fail-on", "debug"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (usage error)", code)
	}
	if !strings.Contains(stderr.String(), "fail-on must be one of") {
		t.Fatalf("stderr = %q, want allowed severities", stderr.String())
	}
}

func TestRunDoctorBaselineSuppressesFindings(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	baselinePath := filepath.Join(t.TempDir(), "doctor-baseline.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(
		[]string{"doctor", "tests/fixtures/doctor/project", "--write-baseline", baselinePath},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("write baseline exit code = %d, stderr = %q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{
			"doctor",
			"tests/fixtures/doctor/project",
			"--baseline",
			baselinePath,
			"--fail-on",
			"warning",
		},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("baseline exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Findings: 0") {
		t.Fatalf("stdout = %q, want suppressed findings", stdout.String())
	}
}

func TestRunDoctorUsesDefaultIgnoreFile(t *testing.T) {
	project := t.TempDir()
	appDir := filepath.Join(project, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("create app dir: %v", err)
	}
	mainPath := filepath.Join(appDir, "main.py")
	writes := map[string]string{
		filepath.Join(project, ".env"):         "",
		filepath.Join(project, ".env.example"): "API_KEY=\n",
		mainPath:                               "import os\nos.environ[\"API_KEY\"]\n",
	}
	for filePath, content := range writes {
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", filePath, err)
		}
	}
	// Suppression keys use the scanned (symlink-resolved) usage path, so build
	// the ignore key from the resolved main.py path.
	resolvedMain, err := filepath.EvalSymlinks(mainPath)
	if err != nil {
		t.Fatalf("resolve main.py: %v", err)
	}
	ignoreContent := "missing:" + resolvedMain + ":API_KEY\n"
	if err := os.WriteFile(filepath.Join(project, ".envdiffignore"), []byte(ignoreContent), 0o644); err != nil {
		t.Fatalf("write .envdiffignore: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"doctor", project, "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, `"ignore_file":`) {
		t.Fatalf("stdout missing ignore_file input:\n%s", out)
	}
	if !strings.Contains(out, ".envdiffignore") {
		t.Fatalf("ignore_file should resolve to the default .envdiffignore:\n%s", out)
	}
	if !strings.Contains(out, `"error": 0`) {
		t.Fatalf("default ignore file should suppress the ENV001 error:\n%s", out)
	}
}
