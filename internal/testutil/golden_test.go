package testutil

import (
	"path/filepath"
	"testing"
)

func TestLoadGoldenJSONLoadsCommittedOracle(t *testing.T) {
	payload := LoadGoldenJSON(t, "compare-basic.json")

	envelope, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("golden payload type = %T, want map[string]any", payload)
	}
	meta, ok := envelope["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta type = %T, want map[string]any", envelope["meta"])
	}
	if meta["command"] != "compare" {
		t.Fatalf("command = %v, want compare", meta["command"])
	}
}

func TestNormalizeJSONValueReplacesRepoRootInStrings(t *testing.T) {
	localPath := filepath.Join(RepoRoot(t), "tests", "fixtures", "repos", "simple_repo")
	payload := map[string]any{
		"root_path": localPath,
		"notes":     []any{"env:" + localPath},
	}

	normalized := NormalizeJSONValue(payload, DefaultPathReplacements(t))

	want := map[string]any{
		"root_path": "<REPO_ROOT>/tests/fixtures/repos/simple_repo",
		"notes":     []any{"env:<REPO_ROOT>/tests/fixtures/repos/simple_repo"},
	}
	AssertJSONEqual(t, normalized, want)
}
