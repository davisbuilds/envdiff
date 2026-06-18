package parsers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanComposeFileDetectsRequiredAndDefaultedValues(t *testing.T) {
	result, err := ScanComposeFile(testutil.FixturePath(t, "compose", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("scan Compose file: %v", err)
	}

	names := []string{}
	for _, usage := range result.Usages {
		names = append(names, usage.Name)
	}
	wantNames := []string{"DATABASE_URL", "DEBUG"}
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
}
