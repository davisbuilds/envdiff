package parsers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanPythonFileDetectsRequiredAndOptionalUsage(t *testing.T) {
	result, err := ScanPythonFile(testutil.FixturePath(t, "python", "sample_app.py"))
	if err != nil {
		t.Fatalf("scan Python file: %v", err)
	}

	names := []string{}
	for _, usage := range result.Usages {
		names = append(names, usage.Name)
	}
	wantNames := []string{"DATABASE_URL", "DEBUG", "REDIS_URL"}
	for index, want := range wantNames {
		if names[index] != want {
			t.Fatalf("names = %v, want %v", names, wantNames)
		}
	}
	if result.Usages[0].Requiredness != "required" {
		t.Fatalf("DATABASE_URL requiredness = %s, want required", result.Usages[0].Requiredness)
	}
	if result.Usages[1].Requiredness != "optional_with_default" {
		t.Fatalf("DEBUG requiredness = %s, want optional_with_default", result.Usages[1].Requiredness)
	}
	if result.Usages[1].DefaultValue == nil || *result.Usages[1].DefaultValue != "false" {
		t.Fatalf("DEBUG default = %v, want false", result.Usages[1].DefaultValue)
	}
	if result.Usages[2].Requiredness != "optional" {
		t.Fatalf("REDIS_URL requiredness = %s, want optional", result.Usages[2].Requiredness)
	}
}

func TestScanPythonFileIgnoresDynamicNames(t *testing.T) {
	result, err := ScanPythonFile(testutil.FixturePath(t, "python", "unsupported.py"))
	if err != nil {
		t.Fatalf("scan Python file: %v", err)
	}

	if len(result.Usages) != 0 {
		t.Fatalf("usages = %#v, want none", result.Usages)
	}
}
