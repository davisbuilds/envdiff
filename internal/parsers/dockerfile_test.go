package parsers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanDockerfileDetectsUsageAndSuppressesArgAndEnv(t *testing.T) {
	result, err := ScanDockerfile(testutil.FixturePath(t, "dockerfile", "Dockerfile"))
	if err != nil {
		t.Fatalf("scan Dockerfile: %v", err)
	}

	usages := map[string]model.EnvVarUsage{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage
		if usage.SourceType != "dockerfile" {
			t.Fatalf("%s source_type = %s, want dockerfile", usage.Name, usage.SourceType)
		}
		if usage.UsageKind != "dockerfile_var" {
			t.Fatalf("%s usage_kind = %s, want dockerfile_var", usage.Name, usage.UsageKind)
		}
	}

	// ARG/ENV locally define these in the same file, so their interpolations are
	// suppressed (NODE_VERSION via ARG, APP_HOME/CACHE_TTL/LOG_LEVEL via ENV).
	for _, suppressed := range []string{"NODE_VERSION", "APP_HOME", "CACHE_TTL", "LOG_LEVEL"} {
		if _, ok := usages[suppressed]; ok {
			t.Fatalf("%s should be suppressed (locally defined)", suppressed)
		}
	}

	want := []string{"API_TOKEN", "TARGET_ENV"}
	if len(usages) != len(want) {
		t.Fatalf("usages = %v, want %v", keysOf(usages), want)
	}
	if usages["TARGET_ENV"].Requiredness != "required" {
		t.Fatalf("TARGET_ENV requiredness = %s, want required", usages["TARGET_ENV"].Requiredness)
	}
	if usages["API_TOKEN"].Requiredness != "required" {
		t.Fatalf("API_TOKEN requiredness = %s, want required", usages["API_TOKEN"].Requiredness)
	}
}
