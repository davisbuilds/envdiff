package analyzers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

var matrixFixturePaths = []string{
	"tests/fixtures/matrix/a.env",
	"tests/fixtures/matrix/b.env",
	"tests/fixtures/matrix/c.env",
}

func TestMatrixDotenvFilesReportsOnlyInconsistentVariablesByDefault(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	result, err := MatrixDotenvFiles(matrixFixturePaths, false)
	if err != nil {
		t.Fatalf("matrix dotenv files: %v", err)
	}

	if result["file_count"] != 3 {
		t.Fatalf("file_count = %v, want 3", result["file_count"])
	}
	if result["variable_count"] != 4 {
		t.Fatalf("variable_count = %v, want 4", result["variable_count"])
	}
	if result["inconsistent_variable_count"] != 3 {
		t.Fatalf("inconsistent_variable_count = %v, want 3", result["inconsistent_variable_count"])
	}
	variables := result["variables"].([]map[string]any)
	names := []string{}
	for _, variable := range variables {
		names = append(names, variable["name"].(string))
	}
	want := []string{"API_KEY", "DATABASE_URL", "DEBUG"}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}

func TestMatrixDotenvFilesShowAllIncludesConsistentVariables(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	result, err := MatrixDotenvFiles(matrixFixturePaths, true)
	if err != nil {
		t.Fatalf("matrix dotenv files: %v", err)
	}

	variables := result["variables"].([]map[string]any)
	names := []string{}
	for _, variable := range variables {
		names = append(names, variable["name"].(string))
	}
	want := []string{"API_KEY", "DATABASE_URL", "DEBUG", "SHARED_MODE"}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}

func TestMatrixDotenvFilesMatchesGoldens(t *testing.T) {
	assertMatrixMatchesGolden(t, false, "matrix-basic.json")
	assertMatrixMatchesGolden(t, true, "matrix-show-all.json")
}

func assertMatrixMatchesGolden(t *testing.T, showAll bool, golden string) {
	t.Helper()
	t.Chdir(testutil.RepoRoot(t))

	result, err := MatrixDotenvFiles(matrixFixturePaths, showAll)
	if err != nil {
		t.Fatalf("matrix dotenv files: %v", err)
	}
	envelope := model.NewJsonEnvelope(
		"matrix",
		map[string]any{"paths": matrixFixturePaths, "show_all": showAll},
		result,
	)
	rendered, err := render.JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	got := testutil.DecodeJSON(t, []byte(rendered))
	testutil.AssertGoldenJSON(t, golden, got)
}
