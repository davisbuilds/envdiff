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
