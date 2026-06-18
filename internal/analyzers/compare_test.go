package analyzers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestCompareDotenvFilesReportsMissingDuplicatesAndKindChanges(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	result, err := CompareDotenvFiles(
		"tests/fixtures/compare/left.env",
		"tests/fixtures/compare/right.env",
	)
	if err != nil {
		t.Fatalf("compare dotenv files: %v", err)
	}

	if result["missing_in_left"].([]string)[0] != "FEATURE" {
		t.Fatalf("missing_in_left = %#v, want FEATURE", result["missing_in_left"])
	}
	if result["missing_in_right"].([]string)[0] != "DUP_KEY" {
		t.Fatalf("missing_in_right = %#v, want DUP_KEY", result["missing_in_right"])
	}
	if result["duplicates_in_left"].([]string)[0] != "DUP_KEY" {
		t.Fatalf("duplicates_in_left = %#v, want DUP_KEY", result["duplicates_in_left"])
	}
	differing := result["differing_values"].([]map[string]any)
	if differing[0]["name"] != "DATABASE_URL" {
		t.Fatalf("differing_values = %#v, want DATABASE_URL first", differing)
	}
}

func TestCompareDotenvFilesMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	result, err := CompareDotenvFiles(
		"tests/fixtures/compare/left.env",
		"tests/fixtures/compare/right.env",
	)
	if err != nil {
		t.Fatalf("compare dotenv files: %v", err)
	}
	envelope := model.NewJsonEnvelope(
		"compare",
		map[string]any{
			"left":  "tests/fixtures/compare/left.env",
			"right": "tests/fixtures/compare/right.env",
		},
		result,
	)
	rendered, err := render.JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	got := testutil.DecodeJSON(t, []byte(rendered))
	want := testutil.LoadGoldenJSON(t, "compare-basic.json")
	testutil.AssertJSONEqual(t, got, want)
}
