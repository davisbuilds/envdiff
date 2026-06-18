package analyzers

import (
	"path/filepath"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
)

func TestBuildBaselineSnapshotOrdersEntriesBySuppressionKey(t *testing.T) {
	findings := []model.Finding{
		{Code: "ENV003", Severity: "info", Title: "Unused", SuppressionKey: model.StringPtr("z")},
		{Code: "ENV001", Severity: "error", Title: "Missing", SuppressionKey: model.StringPtr("a")},
		{Code: "ENV002", Severity: "warning", Title: "No key"},
	}

	snapshot := BuildBaselineSnapshot(findings)

	if len(snapshot.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(snapshot.Entries))
	}
	if snapshot.Entries[0].SuppressionKey != "a" || snapshot.Entries[1].SuppressionKey != "z" {
		t.Fatalf("entries = %#v, want sorted by key", snapshot.Entries)
	}
}

func TestWriteLoadAndApplySuppressions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	findings := []model.Finding{
		{Code: "ENV001", Severity: "error", Title: "Missing", SuppressionKey: model.StringPtr("missing:a")},
		{Code: "ENV003", Severity: "info", Title: "Unused", SuppressionKey: model.StringPtr("unused:b")},
	}

	if _, err := WriteBaselineSnapshot(path, findings); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	snapshot, err := LoadBaselineSnapshot(path)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}
	keys := map[string]struct{}{}
	for _, entry := range snapshot.Entries {
		keys[entry.SuppressionKey] = struct{}{}
	}
	active, suppressed := ApplySuppressions(findings, keys)

	if len(active) != 0 {
		t.Fatalf("active count = %d, want 0", len(active))
	}
	if len(suppressed) != 2 {
		t.Fatalf("suppressed count = %d, want 2", len(suppressed))
	}
}
