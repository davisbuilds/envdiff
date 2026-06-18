package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestScanGitHubActionsFileExtractsSecretAndVarReferences(t *testing.T) {
	result, err := ScanGitHubActionsFile(testutil.FixturePath(t, "github_actions", "deploy.yml"))
	if err != nil {
		t.Fatalf("scan GitHub Actions file: %v", err)
	}

	usages := map[string]string{}
	defaults := map[string]*string{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage.Requiredness
		defaults[usage.Name] = usage.DefaultValue
		if usage.SourceType != "github_actions" {
			t.Fatalf("source type = %s, want github_actions", usage.SourceType)
		}
	}

	if usages["API_KEY"] != "required" {
		t.Fatalf("API_KEY requiredness = %s, want required", usages["API_KEY"])
	}
	if usages["DATABASE_URL"] != "required" {
		t.Fatalf("DATABASE_URL requiredness = %s, want required", usages["DATABASE_URL"])
	}
	if usages["DEPLOY_ENV"] != "optional_with_default" {
		t.Fatalf("DEPLOY_ENV requiredness = %s, want optional_with_default", usages["DEPLOY_ENV"])
	}
	if defaults["DEPLOY_ENV"] == nil || *defaults["DEPLOY_ENV"] != "staging" {
		t.Fatalf("DEPLOY_ENV default = %v, want staging", defaults["DEPLOY_ENV"])
	}
}

func TestScanGitHubActionsFileHandlesBlankAndUnquotedDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deploy.yml")
	content := "env:\n  EMPTY: ${{ vars.EMPTY || }}\n  REGION: ${{ vars.REGION || us-east-1 }}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	result, err := ScanGitHubActionsFile(path)
	if err != nil {
		t.Fatalf("scan GitHub Actions file: %v", err)
	}

	usages := map[string]string{}
	defaults := map[string]*string{}
	for _, usage := range result.Usages {
		usages[usage.Name] = usage.Requiredness
		defaults[usage.Name] = usage.DefaultValue
	}

	if usages["EMPTY"] != "required" {
		t.Fatalf("EMPTY requiredness = %s, want required", usages["EMPTY"])
	}
	if defaults["EMPTY"] != nil {
		t.Fatalf("EMPTY default = %v, want nil", *defaults["EMPTY"])
	}
	if usages["REGION"] != "optional_with_default" {
		t.Fatalf("REGION requiredness = %s, want optional_with_default", usages["REGION"])
	}
	if defaults["REGION"] == nil || *defaults["REGION"] != "us-east-1" {
		t.Fatalf("REGION default = %v, want us-east-1", defaults["REGION"])
	}
}
