package render

import (
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
)

func TestScanResultIncludesCountsAndContracts(t *testing.T) {
	result := model.RepoScanResult{
		RootPath:    "/repo",
		Definitions: []model.EnvVarDefinition{{Name: "DATABASE_URL"}},
		Usages:      []model.EnvVarUsage{{Name: "DATABASE_URL"}},
		Contracts:   []model.EnvVarContract{{Name: "DATABASE_URL", Requiredness: "required", Status: []string{"defined", "referenced"}}},
		Resolutions: []model.ResolutionDecision{},
		Warnings:    []string{},
	}

	rendered := ScanResult(result)

	for _, want := range []string{"Scan root: /repo", "Contracts: 1", "DATABASE_URL [required]"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered scan missing %q:\n%s", want, rendered)
		}
	}
}

func TestDoctorResultGroupsFindingsBySeverity(t *testing.T) {
	findings := []model.Finding{
		{Code: "ENV001", Severity: "error", Details: "missing"},
		{Code: "ENV002", Severity: "warning", Details: "optional missing"},
		{Code: "ENV003", Severity: "info", Details: "unused"},
	}

	rendered := DoctorResult("/repo", findings, 1, nil)

	for _, want := range []string{"Summary: 1 error, 1 warning, 1 info", "Suppressed: 1", "Errors:", "Warnings:", "Infos:"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered doctor missing %q:\n%s", want, rendered)
		}
	}
}

func TestGenerateResultReportsCheckAndWriteModes(t *testing.T) {
	outputPath := "/repo/.env.example"
	matches := true

	if got := GenerateResult(3, &outputPath, true, nil, nil); got != "Generated 3 variables with annotations to /repo/.env.example" {
		t.Fatalf("write result = %q", got)
	}
	if got := GenerateResult(3, nil, false, &outputPath, &matches); got != "Generated output matches /repo/.env.example" {
		t.Fatalf("check result = %q", got)
	}
}
