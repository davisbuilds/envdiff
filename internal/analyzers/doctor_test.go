package analyzers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestDoctorRepositoryEmitsCoreFindings(t *testing.T) {
	scanResult, err := ScanRepository(testutil.FixturePath(t, "doctor", "project"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	findings := DoctorRepository(scanResult)
	codes := map[string]struct{}{}
	for _, finding := range findings {
		codes[finding.Code] = struct{}{}
	}
	for _, code := range []string{"ENV001", "ENV002", "ENV003", "ENV004", "ENV006", "ENV007", "ENV008", "ENV009"} {
		if _, ok := codes[code]; !ok {
			t.Fatalf("missing finding code %s in %#v", code, findings)
		}
	}
}

func TestDoctorRepositoryMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	scanResult, err := ScanRepository("tests/fixtures/doctor/project")
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	findings := DoctorRepository(scanResult)
	envelope := model.NewJsonEnvelope(
		"doctor",
		map[string]any{
			"baseline":       nil,
			"fail_on":        "error",
			"ignore_file":    nil,
			"path":           "tests/fixtures/doctor/project",
			"write_baseline": nil,
		},
		map[string]any{
			"filtering": map[string]any{
				"baseline_entries": 0,
				"baseline_written": nil,
				"suppressed_count": 0,
			},
			"scan":                scanResult,
			"suppressed_findings": []model.Finding{},
		},
	)
	envelope.Findings = findings
	envelope.Summary = SummarizeFindings(findings)

	rendered, err := render.JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	got := testutil.DecodeJSON(t, []byte(rendered))
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, "doctor-project.json")
	testutil.AssertJSONEqual(t, got, want)
}
