package order

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
)

func TestDefinitionsSortByNamePathAndLine(t *testing.T) {
	definitions := []model.EnvVarDefinition{
		{Name: "B", FilePath: "a.env", LineNumber: 1},
		{Name: "A", FilePath: "b.env", LineNumber: 2},
		{Name: "A", FilePath: "a.env", LineNumber: 3},
		{Name: "A", FilePath: "a.env", LineNumber: 1},
	}

	ordered := Definitions(definitions)

	got := []int{ordered[0].LineNumber, ordered[1].LineNumber, ordered[2].LineNumber, ordered[3].LineNumber}
	want := []int{1, 3, 2, 1}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("line order = %v, want %v", got, want)
		}
	}
}

func TestUsagesSortNilLineNumbersBeforeConcreteLines(t *testing.T) {
	line := 4
	usages := []model.EnvVarUsage{
		{Name: "DATABASE_URL", FilePath: "app.py", LineNumber: &line},
		{Name: "DATABASE_URL", FilePath: "app.py"},
	}

	ordered := Usages(usages)

	if ordered[0].LineNumber != nil {
		t.Fatalf("first line number = %v, want nil", *ordered[0].LineNumber)
	}
}

func TestFindingsSortBySeverityCodeVariableAndLocation(t *testing.T) {
	warningVar := "A"
	errorVar := "B"
	findings := []model.Finding{
		{Code: "ENV003", Severity: "info", VariableName: &warningVar},
		{Code: "ENV002", Severity: "warning", VariableName: &warningVar},
		{
			Code:         "ENV001",
			Severity:     "error",
			VariableName: &errorVar,
			Locations:    []model.Location{{FilePath: "b.py", LineNumber: model.IntPtr(1)}},
		},
		{
			Code:         "ENV001",
			Severity:     "error",
			VariableName: &errorVar,
			Locations:    []model.Location{{FilePath: "a.py", LineNumber: model.IntPtr(1)}},
		},
	}

	ordered := Findings(findings)

	if ordered[0].Locations[0].FilePath != "a.py" {
		t.Fatalf("first finding path = %s, want a.py", ordered[0].Locations[0].FilePath)
	}
	if ordered[1].Locations[0].FilePath != "b.py" {
		t.Fatalf("second finding path = %s, want b.py", ordered[1].Locations[0].FilePath)
	}
	if ordered[2].Severity != "warning" {
		t.Fatalf("third severity = %s, want warning", ordered[2].Severity)
	}
	if ordered[3].Severity != "info" {
		t.Fatalf("fourth severity = %s, want info", ordered[3].Severity)
	}
}
