package analyzers

import (
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanRepositoryBuildsContractsForSimpleRepo(t *testing.T) {
	result, err := ScanRepository(testutil.FixturePath(t, "repos", "simple_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	names := make([]string, 0, len(result.Contracts))
	for _, contract := range result.Contracts {
		names = append(names, contract.Name)
	}
	want := []string{"DATABASE_URL", "DEBUG", "REDIS_URL"}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("contracts = %v, want %v", names, want)
		}
	}
	if len(result.Resolutions) != 1 {
		t.Fatalf("resolution count = %d, want 1", len(result.Resolutions))
	}
	if result.Resolutions[0].EnvFile == nil {
		t.Fatal("resolution env file = nil, want .env path")
	}
}

func TestScanRepositoryMatchesSimpleRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/simple_repo", "scan-simple-repo.json")
}

func TestScanRepositoryMatchesWorkflowRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/workflow_repo", "scan-workflow-repo.json")
}

func TestScanRepositoryMatchesUnicodeRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/unicode_repo", "scan-unicode-repo.json")
}

// Go is the source of truth for the JSON contract, which emits non-ASCII values
// as raw UTF-8 rather than \uXXXX escapes (unlike Python's ensure_ascii default).
// This pins that behavior at the byte level, where structural golden comparison
// (which decodes both sides) cannot see the difference.
func TestScanRepositoryEmitsRawUTF8(t *testing.T) {
	const greeting = "Café ☕ déjà vu — Москва"

	result, err := ScanRepository(testutil.FixturePath(t, "repos", "unicode_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	rendered, err := render.JSON(model.NewJsonEnvelope("scan", map[string]any{}, result))
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	if !strings.Contains(rendered, greeting) {
		t.Fatalf("rendered JSON missing raw UTF-8 value %q:\n%s", greeting, rendered)
	}
	if strings.Contains(rendered, `\u`) {
		t.Fatalf("rendered JSON contains \\u escapes, want raw UTF-8:\n%s", rendered)
	}
}

func assertScanMatchesGolden(t *testing.T, fixture string, golden string) {
	t.Helper()

	result, err := ScanRepository(testutil.FixturePath(t, filepathParts(fixture)...))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	envelope := model.NewJsonEnvelope(
		"scan",
		map[string]any{"path": "tests/fixtures/" + fixture},
		result,
	)
	rendered, err := render.JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	got := testutil.DecodeJSON(t, []byte(rendered))
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, golden)
	testutil.AssertJSONEqual(t, got, want)
}

func filepathParts(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
