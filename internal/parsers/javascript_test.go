package parsers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanJavaScriptFileDetectsProcessEnvAccess(t *testing.T) {
	result, err := ScanJavaScriptFile(testutil.FixturePath(t, "javascript", "sample_app.ts"))
	if err != nil {
		t.Fatalf("scan JS file: %v", err)
	}

	usages := map[string]model.EnvVarUsage{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage
		if usage.SourceType != "javascript" {
			t.Fatalf("%s source_type = %s, want javascript", usage.Name, usage.SourceType)
		}
		if usage.UsageKind != "process.env" {
			t.Fatalf("%s usage_kind = %s, want process.env", usage.Name, usage.UsageKind)
		}
	}

	want := []string{"AWS_REGION", "DATABASE_URL", "DEBUG", "LOG_LEVEL", "PORT"}
	if len(usages) != len(want) {
		t.Fatalf("usages = %v, want %v", keysOf(usages), want)
	}

	// Bare reads return undefined in JS (no throw), so they are optional.
	if usages["DATABASE_URL"].Requiredness != "optional" {
		t.Fatalf("DATABASE_URL requiredness = %s, want optional", usages["DATABASE_URL"].Requiredness)
	}
	if usages["DATABASE_URL"].DefaultValue != nil {
		t.Fatalf("DATABASE_URL default = %v, want nil", *usages["DATABASE_URL"].DefaultValue)
	}
	if usages["AWS_REGION"].Requiredness != "optional" {
		t.Fatalf("AWS_REGION requiredness = %s, want optional", usages["AWS_REGION"].Requiredness)
	}

	// Inline || / ?? fallbacks are optional_with_default.
	assertDefault(t, usages["PORT"], "3000")
	assertDefault(t, usages["LOG_LEVEL"], "info")
	assertDefault(t, usages["DEBUG"], "false")
}

func TestScanJavaScriptFileDetectsViteImportMetaEnv(t *testing.T) {
	result, err := ScanJavaScriptFile(testutil.FixturePath(t, "javascript", "vite_app.ts"))
	if err != nil {
		t.Fatalf("scan JS file: %v", err)
	}

	usages := map[string]model.EnvVarUsage{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage
		if usage.UsageKind != "import.meta.env" {
			t.Fatalf("%s usage_kind = %s, want import.meta.env", usage.Name, usage.UsageKind)
		}
	}

	want := []string{"MODE", "VITE_API_URL", "VITE_FEATURE_FLAG"}
	if len(usages) != len(want) {
		t.Fatalf("usages = %v, want %v", keysOf(usages), want)
	}
	if usages["VITE_API_URL"].Requiredness != "optional" {
		t.Fatalf("VITE_API_URL requiredness = %s, want optional", usages["VITE_API_URL"].Requiredness)
	}
	assertDefault(t, usages["MODE"], "development")
}

func TestScanJavaScriptFileDetectsDestructuring(t *testing.T) {
	result, err := ScanJavaScriptFile(testutil.FixturePath(t, "javascript", "destructure.js"))
	if err != nil {
		t.Fatalf("scan JS file: %v", err)
	}

	usages := map[string]model.EnvVarUsage{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage
		if usage.UsageKind != "process.env" {
			t.Fatalf("%s usage_kind = %s, want process.env", usage.Name, usage.UsageKind)
		}
		if usage.Requiredness != "optional" {
			t.Fatalf("%s requiredness = %s, want optional", usage.Name, usage.Requiredness)
		}
	}

	want := []string{"API_URL", "AUTH_TOKEN", "NODE_ENV"}
	if len(usages) != len(want) {
		t.Fatalf("usages = %v, want %v (renamed key takes the source name)", keysOf(usages), want)
	}
}

func TestScanJavaScriptFileIgnoresDynamicAndWholeObjectAccess(t *testing.T) {
	result, err := ScanJavaScriptFile(testutil.FixturePath(t, "javascript", "unsupported.js"))
	if err != nil {
		t.Fatalf("scan JS file: %v", err)
	}
	if len(result.Usages) != 0 {
		t.Fatalf("usages = %#v, want none", result.Usages)
	}
}

// assertDefault checks an optional_with_default usage and its captured default.
func assertDefault(t *testing.T, usage model.EnvVarUsage, wantDefault string) {
	t.Helper()
	if usage.Requiredness != "optional_with_default" {
		t.Fatalf("%s requiredness = %s, want optional_with_default", usage.Name, usage.Requiredness)
	}
	if usage.DefaultValue == nil || *usage.DefaultValue != wantDefault {
		t.Fatalf("%s default = %v, want %s", usage.Name, usage.DefaultValue, wantDefault)
	}
}

func keysOf(m map[string]model.EnvVarUsage) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
