package render

import (
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
)

func TestCompareResultRendersMissingDuplicatesAndDifferingValues(t *testing.T) {
	rendered := CompareResult(map[string]any{
		"left_path":           ".env",
		"right_path":          ".env.example",
		"missing_in_left":     []string{"API_KEY"},
		"missing_in_right":    []string{},
		"duplicates_in_left":  []string{"PORT"},
		"duplicates_in_right": []string{},
		"differing_values": []map[string]any{
			{
				"name":        "DEBUG",
				"left_kind":   "literal",
				"left_value":  "true",
				"right_kind":  "placeholder",
				"right_value": "<debug>",
			},
		},
	})

	want := strings.Join([]string{
		"Compare .env vs .env.example",
		"Missing in left: API_KEY",
		"Missing in right: -",
		"Duplicates in left: PORT",
		"Duplicates in right: -",
		"Differing values:",
		"  DEBUG: literal=true vs placeholder=<debug>",
	}, "\n")
	if rendered != want {
		t.Fatalf("compare render =\n%s\n\nwant\n%s", rendered, want)
	}
}

func TestScanResultRendersContractSummaryWithStatuses(t *testing.T) {
	rendered := ScanResult(model.RepoScanResult{
		RootPath:    "/repo",
		Definitions: []model.EnvVarDefinition{{Name: "DATABASE_URL"}},
		Usages:      []model.EnvVarUsage{{Name: "DATABASE_URL"}},
		Contracts: []model.EnvVarContract{
			{Name: "DATABASE_URL", Requiredness: "required", Status: []string{"defined", "used"}},
			{Name: "OPTIONAL_FLAG", Requiredness: "optional"},
		},
	})

	want := strings.Join([]string{
		"Scan root: /repo",
		"Definitions: 1",
		"Usages: 1",
		"Contracts: 2",
		"Contracts:",
		"  DATABASE_URL [required] (defined,used)",
		"  OPTIONAL_FLAG [optional] (none)",
	}, "\n")
	if rendered != want {
		t.Fatalf("scan render =\n%s\n\nwant\n%s", rendered, want)
	}
}

func TestMatrixResultRendersEmptyViewWithWarningCount(t *testing.T) {
	rendered := MatrixResult(map[string]any{
		"file_count":                  2,
		"variable_count":              0,
		"inconsistent_variable_count": 0,
		"warnings":                    []string{"ignored file"},
		"variables":                   []map[string]any{},
	})

	want := strings.Join([]string{
		"Matrix files: 2",
		"Variables: 0",
		"Inconsistent: 0",
		"Warnings: 1",
		"No variables matched the selected view.",
	}, "\n")
	if rendered != want {
		t.Fatalf("matrix render =\n%s\n\nwant\n%s", rendered, want)
	}
}

func TestMatrixResultRendersVariableReasonsAndFilePresence(t *testing.T) {
	rendered := MatrixResult(map[string]any{
		"file_count":                  2,
		"variable_count":              2,
		"inconsistent_variable_count": 1,
		"warnings":                    []string{},
		"variables": []map[string]any{
			{
				"name":          "DATABASE_URL",
				"missing_in":    []string{".env.example"},
				"value_kinds":   []string{"literal", "placeholder"},
				"duplicates_in": []string{".env"},
				"files": []map[string]any{
					{"path": ".env", "presence": "present", "value_kind": "literal", "is_duplicate": true},
					{"path": ".env.example", "presence": "missing"},
				},
			},
			{
				"name":          "LOG_LEVEL",
				"missing_in":    []string{},
				"value_kinds":   []string{"literal"},
				"duplicates_in": []string{},
				"files": []map[string]any{
					{"path": ".env", "presence": "present", "value_kind": "literal", "is_duplicate": false},
				},
			},
		},
	})

	want := strings.Join([]string{
		"Matrix files: 2",
		"Variables: 2",
		"Inconsistent: 1",
		"Variables:",
		"  DATABASE_URL [missing, kind-mismatch, duplicate]",
		"    .env: present (literal, duplicate)",
		"    .env.example: missing",
		"  LOG_LEVEL [consistent]",
		"    .env: present (literal)",
	}, "\n")
	if rendered != want {
		t.Fatalf("matrix render =\n%s\n\nwant\n%s", rendered, want)
	}
}

func TestGenerateResultRendersCheckAndWriteModes(t *testing.T) {
	checkPath := ".env.example"
	matches := true
	drifted := false

	if got := GenerateResult(3, nil, false, &checkPath, &matches); got != "Generated output matches .env.example" {
		t.Fatalf("matches result = %q", got)
	}
	if got := GenerateResult(3, nil, false, &checkPath, &drifted); got != "Generated output drifted from .env.example" {
		t.Fatalf("drift result = %q", got)
	}
	outputPath := ".env.example"
	if got := GenerateResult(3, &outputPath, true, nil, nil); got != "Generated 3 variables with annotations to .env.example" {
		t.Fatalf("write result = %q", got)
	}
}

func TestDoctorResultRendersEmptyFindingsWithSuppressionMetadata(t *testing.T) {
	baselineWritten := ".envdiff-baseline.json"
	rendered := DoctorResult("/repo", nil, 2, &baselineWritten)

	want := strings.Join([]string{
		"Doctor root: /repo",
		"Findings: 0",
		"Summary: 0 error, 0 warning, 0 info",
		"Suppressed: 2",
		"Baseline written: .envdiff-baseline.json",
		"No active findings.",
	}, "\n")
	if rendered != want {
		t.Fatalf("doctor render =\n%s\n\nwant\n%s", rendered, want)
	}
}

func TestDoctorResultGroupsFindingsBySeverity(t *testing.T) {
	rendered := DoctorResult("/repo", []model.Finding{
		{Code: "ENV001", Severity: "warning", Title: "Missing definition", Details: "DATABASE_URL is used but not defined"},
		{Code: "ENV002", Severity: "error", Title: "Duplicate definition", Details: "PORT is defined twice"},
	}, 0, nil)

	want := strings.Join([]string{
		"Doctor root: /repo",
		"Findings: 2",
		"Summary: 1 error, 1 warning, 0 info",
		"Errors:",
		"  ENV002 PORT is defined twice",
		"Warnings:",
		"  ENV001 DATABASE_URL is used but not defined",
	}, "\n")
	if rendered != want {
		t.Fatalf("doctor render =\n%s\n\nwant\n%s", rendered, want)
	}
}
