package analyzers

import (
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
)

func TestSecretAndPlaceholderFindingsOnlyInspectCommittedEnv(t *testing.T) {
	scanResult := model.RepoScanResult{
		Definitions: []model.EnvVarDefinition{
			{
				FilePath:            "/repo/.env",
				LineNumber:          1,
				Name:                "OPENAI_KEY",
				NormalizedValueKind: "secret_like",
				Value:               "sk-proj-1234567890abcdef1234567890",
			},
			{
				FilePath:            "/repo/.env",
				LineNumber:          2,
				Name:                "PLACEHOLDER",
				NormalizedValueKind: "placeholder",
				Value:               "changeme",
			},
			{
				FilePath:            "/repo/.env.example",
				LineNumber:          1,
				Name:                "EXAMPLE_PLACEHOLDER",
				NormalizedValueKind: "placeholder",
				Value:               "changeme",
			},
		},
	}

	findings := SecretAndPlaceholderFindings(scanResult)

	codes := []string{}
	for _, finding := range findings {
		codes = append(codes, finding.Code)
	}
	if len(codes) != 2 || codes[0] != "ENV008" || codes[1] != "ENV009" {
		t.Fatalf("codes = %v, want ENV008 and ENV009", codes)
	}
}
