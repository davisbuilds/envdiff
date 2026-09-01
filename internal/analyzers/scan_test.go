package analyzers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanRepositoryBuildsContractsForSimpleRepo(t *testing.T) {
	result, err := ScanRepository(testutil.FixturePath(t, "repos", "simple_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	names := make([]string, 0, len(result.Contracts))
	for _, contract := range result.Contracts {
		names = append(names, contract.Name)
	}
	want := []string{"DATABASE_URL", "DEBUG", "REDIS_URL"}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("contracts = %v, want %v", names, want)
		}
	}
	if len(result.Resolutions) != 1 {
		t.Fatalf("resolution count = %d, want 1", len(result.Resolutions))
	}
	if result.Resolutions[0].EnvFile == nil {
		t.Fatal("resolution env file = nil, want .env path")
	}
}

func TestScanRepositoryMatchesSimpleRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/simple_repo", "scan-simple-repo.json")
}

func TestScanRepositoryMatchesWorkflowRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/workflow_repo", "scan-workflow-repo.json")
}

func TestScanRepositoryMatchesUnicodeRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/unicode_repo", "scan-unicode-repo.json")
}

func TestScanRepositoryMatchesNodeRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/node_repo", "scan-node-repo.json")
}

func TestScanRepositoryMatchesDeployRepoGolden(t *testing.T) {
	assertScanMatchesGolden(t, "repos/deploy_repo", "scan-deploy-repo.json")
}

func TestScanRepositoryDetectsShellAndDockerfileUsage(t *testing.T) {
	result, err := ScanRepository(testutil.FixturePath(t, "repos", "deploy_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	names := map[string]model.EnvVarContract{}
	for _, contract := range result.Contracts {
		names[contract.Name] = contract
	}
	for _, want := range []string{"DATABASE_URL", "REDIS_URL", "SENTRY_DSN"} {
		if _, ok := names[want]; !ok {
			t.Fatalf("contracts = %v, want %s referenced", keysOfContracts(names), want)
		}
	}
	// NODE_ENV is defined locally by the Dockerfile ENV and must not surface as a
	// repo-wide contract (separate scope).
	if _, ok := names["NODE_ENV"]; ok {
		t.Fatalf("NODE_ENV should not be a contract (locally defined in Dockerfile)")
	}
}

func TestScanRepositoryDetectsJavaScriptProcessEnvUsage(t *testing.T) {
	result, err := ScanRepository(testutil.FixturePath(t, "repos", "node_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	names := map[string]model.EnvVarContract{}
	for _, contract := range result.Contracts {
		names[contract.Name] = contract
	}
	for _, want := range []string{"DATABASE_URL", "PORT", "LOG_LEVEL"} {
		if _, ok := names[want]; !ok {
			t.Fatalf("contracts = %v, want %s referenced", keysOfContracts(names), want)
		}
	}
	// PORT is referenced from JS but absent from .env.
	if status := names["PORT"].Status; len(status) == 0 || status[0] != "referenced" {
		t.Fatalf("PORT status = %v, want referenced (undefined)", names["PORT"].Status)
	}
}

func TestScanRepositoryScansDirenvEnvrcAsShell(t *testing.T) {
	root := t.TempDir()
	writeDoctorFile(t, filepath.Join(root, ".env"), "DATABASE_URL=postgres://db\n")
	writeDoctorFile(t, filepath.Join(root, ".envrc"), "export NODE_ENV=development\necho \"$STRIPE_KEY\"\n")

	result, err := ScanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	names := map[string]model.EnvVarContract{}
	for _, contract := range result.Contracts {
		names[contract.Name] = contract
	}
	// STRIPE_KEY is referenced in .envrc (scanned as shell).
	if _, ok := names["STRIPE_KEY"]; !ok {
		t.Fatalf("contracts = %v, want STRIPE_KEY from .envrc", keysOfContracts(names))
	}
	// NODE_ENV is locally exported in the same .envrc, so it is suppressed.
	if _, ok := names["NODE_ENV"]; ok {
		t.Fatalf("NODE_ENV should be suppressed (locally exported in .envrc)")
	}
}

func TestResolutionCacheReusesNearestFilesForSiblingUsageDirectories(t *testing.T) {
	root := t.TempDir()
	envFile := filepath.Join(root, ".env")
	if err := os.WriteFile(envFile, []byte("DATABASE_URL=postgres://db\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	firstDirectory := filepath.Join(root, "services", "api", "handlers")
	secondDirectory := filepath.Join(root, "services", "api", "workers")
	for _, directory := range []string{firstDirectory, secondDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatalf("make usage directory: %v", err)
		}
	}

	cache := newResolutionCache(root)
	first := cache.resolveUsageFile(filepath.Join(firstDirectory, "request.py"))
	if first.EnvFile == nil || *first.EnvFile != envFile {
		t.Fatalf("first env file = %v, want %q", first.EnvFile, envFile)
	}
	if err := os.Remove(envFile); err != nil {
		t.Fatalf("remove .env after initial resolution: %v", err)
	}

	second := cache.resolveUsageFile(filepath.Join(secondDirectory, "worker.py"))
	if second.EnvFile == nil || *second.EnvFile != envFile {
		t.Fatalf("second env file = %v, want cached %q", second.EnvFile, envFile)
	}
}

func keysOfContracts(m map[string]model.EnvVarContract) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// Go is the source of truth for the JSON contract, which emits non-ASCII values
// as raw UTF-8 rather than \uXXXX escapes (unlike Python's ensure_ascii default).
// This pins that behavior at the byte level, where structural golden comparison
// (which decodes both sides) cannot see the difference.
func TestScanRepositoryEmitsRawUTF8(t *testing.T) {
	const greeting = "Café ☕ déjà vu — Москва"

	result, err := ScanRepository(testutil.FixturePath(t, "repos", "unicode_repo"))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	rendered, err := render.JSON(model.NewJsonEnvelope("scan", map[string]any{}, result))
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	if !strings.Contains(rendered, greeting) {
		t.Fatalf("rendered JSON missing raw UTF-8 value %q:\n%s", greeting, rendered)
	}
	if strings.Contains(rendered, `\u`) {
		t.Fatalf("rendered JSON contains \\u escapes, want raw UTF-8:\n%s", rendered)
	}
}

func assertScanMatchesGolden(t *testing.T, fixture string, golden string) {
	t.Helper()

	result, err := ScanRepository(testutil.FixturePath(t, filepathParts(fixture)...))
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	envelope := model.NewJsonEnvelope(
		"scan",
		map[string]any{"path": "tests/fixtures/" + fixture},
		result,
	)
	rendered, err := render.JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	got := testutil.DecodeJSON(t, []byte(rendered))
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	testutil.AssertGoldenJSON(t, golden, got)
}

func filepathParts(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
