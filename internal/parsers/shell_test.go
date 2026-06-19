package parsers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanShellFileDetectsUsageAndSuppressesLocalDefinitions(t *testing.T) {
	result, err := ScanShellFile(testutil.FixturePath(t, "shell", "entrypoint.sh"))
	if err != nil {
		t.Fatalf("scan shell file: %v", err)
	}

	usages := map[string]model.EnvVarUsage{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage
		if usage.SourceType != "shell" {
			t.Fatalf("%s source_type = %s, want shell", usage.Name, usage.SourceType)
		}
		if usage.UsageKind != "shell_var" {
			t.Fatalf("%s usage_kind = %s, want shell_var", usage.Name, usage.UsageKind)
		}
	}

	// PORT (export) and LOG_DIR (bare assignment) are locally defined, so their
	// $PORT/$LOG_DIR references are suppressed.
	for _, suppressed := range []string{"PORT", "LOG_DIR", "region"} {
		if _, ok := usages[suppressed]; ok {
			t.Fatalf("%s should be suppressed (locally defined)", suppressed)
		}
	}

	want := []string{"API_KEY", "AWS_REGION", "DATABASE_URL"}
	if len(usages) != len(want) {
		t.Fatalf("usages = %v, want %v", keysOf(usages), want)
	}

	// Bare interpolation is required (Compose precedent); :? is also required.
	if usages["DATABASE_URL"].Requiredness != "required" {
		t.Fatalf("DATABASE_URL requiredness = %s, want required", usages["DATABASE_URL"].Requiredness)
	}
	if usages["API_KEY"].Requiredness != "required" {
		t.Fatalf("API_KEY requiredness = %s, want required", usages["API_KEY"].Requiredness)
	}
	assertDefault(t, usages["AWS_REGION"], "us-east-1")
}
