package analyzers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestShouldFailIsCaseInsensitive(t *testing.T) {
	summary := model.SummaryCounts{Error: 1}
	for _, threshold := range []string{"error", "ERROR", "Error"} {
		fail, err := ShouldFail(summary, threshold)
		if err != nil {
			t.Fatalf("ShouldFail(%q): unexpected error %v", threshold, err)
		}
		if !fail {
			t.Fatalf("ShouldFail(%q): expected true for one error", threshold)
		}
	}

	if _, err := ShouldFail(summary, "bogus"); err == nil {
		t.Fatalf("ShouldFail(\"bogus\"): expected an error")
	}
}

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
	testutil.AssertGoldenJSON(t, "doctor-project.json", got)
}

func TestDoctorRepositoryDoesNotEmitUnusedForEnvExampleEntries(t *testing.T) {
	scanResult, err := ScanRepository(testutil.FixturePath(t, "doctor", "project"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	for _, finding := range DoctorRepository(scanResult) {
		if finding.Code != "ENV003" {
			continue
		}
		for _, location := range finding.Locations {
			if strings.HasSuffix(location.FilePath, ".env.example") {
				t.Fatalf("ENV003 should not target .env.example: %#v", finding)
			}
		}
	}
}

func TestDoctorRepositoryEmitsTemplateSkewForStaleExampleOnlyVariable(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	writeDoctorFile(t, filepath.Join(root, ".env"), "DATABASE_URL=postgres://localhost/db\n")
	writeDoctorFile(t, filepath.Join(root, ".env.example"), "DATABASE_URL=\nLEGACY_TEMPLATE=\n")
	writeDoctorFile(t, filepath.Join(appDir, "main.py"), "import os\n\ndatabase_url = os.environ[\"DATABASE_URL\"]\n")

	scanResult, err := ScanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	skew := []model.Finding{}
	for _, finding := range DoctorRepository(scanResult) {
		if finding.Code == "ENV005" {
			skew = append(skew, finding)
		}
	}
	if len(skew) != 1 {
		t.Fatalf("ENV005 count = %d, want 1: %#v", len(skew), skew)
	}
	if skew[0].Severity != "info" {
		t.Fatalf("ENV005 severity = %s, want info", skew[0].Severity)
	}
	if skew[0].VariableName == nil || *skew[0].VariableName != "LEGACY_TEMPLATE" {
		t.Fatalf("ENV005 variable = %v, want LEGACY_TEMPLATE", skew[0].VariableName)
	}
}

func TestDoctorRepositoryReportsMissingWorkflowSecretReferences(t *testing.T) {
	scanResult, err := ScanRepository(testutil.FixturePath(t, "repos", "workflow_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	missing := map[string]struct{}{}
	for _, finding := range DoctorRepository(scanResult) {
		if (finding.Code == "ENV001" || finding.Code == "ENV002") && finding.VariableName != nil {
			missing[*finding.VariableName] = struct{}{}
		}
	}
	if len(missing) != 2 {
		t.Fatalf("missing names = %v, want API_KEY and DEPLOY_ENV", missing)
	}
	for _, name := range []string{"API_KEY", "DEPLOY_ENV"} {
		if _, ok := missing[name]; !ok {
			t.Fatalf("missing names = %v, want %s present", missing, name)
		}
	}
}

func TestDoctorRepositoryReportsMissingUsageWithoutAssociatedEnv(t *testing.T) {
	root := t.TempDir()
	writeDoctorFile(t, filepath.Join(root, "main.py"), "import os\nos.environ[\"API_KEY\"]\n")

	scanResult, err := ScanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	var missing *model.Finding
	for _, finding := range DoctorRepository(scanResult) {
		if finding.Code == "ENV001" {
			finding := finding
			missing = &finding
			break
		}
	}
	if missing == nil {
		t.Fatal("expected an ENV001 finding")
	}
	if missing.VariableName == nil || *missing.VariableName != "API_KEY" {
		t.Fatalf("variable = %v, want API_KEY", missing.VariableName)
	}
	if !strings.Contains(missing.Details, "no associated .env defines it") {
		t.Fatalf("details = %q, want no-associated-env phrasing", missing.Details)
	}
}

func TestDoctorRepositoryDeduplicatesAliasFindingsForRepeatedMissingUsage(t *testing.T) {
	root := t.TempDir()
	writeDoctorFile(t, filepath.Join(root, ".env"), "OPENAI_KEY=sk-test\n")
	writeDoctorFile(t, filepath.Join(root, "main.py"),
		"import os\nos.environ[\"OPENAI_API_KEY\"]\nos.environ[\"OPENAI_API_KEY\"]\n")

	scanResult, err := ScanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	alias := []model.Finding{}
	for _, finding := range DoctorRepository(scanResult) {
		if finding.Code == "ENV007" {
			alias = append(alias, finding)
		}
	}
	if len(alias) != 1 {
		t.Fatalf("ENV007 count = %d, want 1 (deduplicated): %#v", len(alias), alias)
	}
	if len(alias[0].RelatedVariables) != 1 || alias[0].RelatedVariables[0] != "OPENAI_KEY" {
		t.Fatalf("related = %v, want [OPENAI_KEY]", alias[0].RelatedVariables)
	}
}

func writeDoctorFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
