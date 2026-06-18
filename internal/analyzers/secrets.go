package analyzers

import (
	"fmt"
	"path/filepath"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/normalize"
	"github.com/davisbuilds/envdiff/internal/order"
)

func SecretAndPlaceholderFindings(scanResult model.RepoScanResult) []model.Finding {
	findings := []model.Finding{}
	seen := map[string]struct{}{}

	for _, definition := range scanResult.Definitions {
		if filepath.Base(definition.FilePath) != ".env" {
			continue
		}

		if definition.NormalizedValueKind == "secret_like" {
			key := fmt.Sprintf("ENV008:%s:%d:%s", definition.FilePath, definition.LineNumber, definition.Name)
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				findings = append(findings, model.Finding{
					Code:       "ENV008",
					Confidence: model.StringPtr("low"),
					Details: fmt.Sprintf(
						"%s in %s looks like a real secret value.",
						definition.Name,
						definition.FilePath,
					),
					Locations: []model.Location{{
						FilePath:   definition.FilePath,
						LineNumber: model.IntPtr(definition.LineNumber),
					}},
					Reason:     model.StringPtr("Value shape is long, opaque, and mixed-character."),
					Severity:   "warning",
					SourceKind: "heuristic",
					SuppressionKey: model.StringPtr(
						fmt.Sprintf(
							"secret-like:%s:%d:%s",
							definition.FilePath,
							definition.LineNumber,
							definition.Name,
						),
					),
					Title:        "Secret-like committed value",
					VariableName: model.StringPtr(definition.Name),
				})
			}
		}

		if normalize.IsNonEmptyPlaceholder(definition.Value) {
			key := fmt.Sprintf("ENV009:%s:%d:%s", definition.FilePath, definition.LineNumber, definition.Name)
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				findings = append(findings, model.Finding{
					Code:       "ENV009",
					Confidence: model.StringPtr("low"),
					Details: fmt.Sprintf(
						"%s in %s uses a placeholder value.",
						definition.Name,
						definition.FilePath,
					),
					Locations: []model.Location{{
						FilePath:   definition.FilePath,
						LineNumber: model.IntPtr(definition.LineNumber),
					}},
					Reason: model.StringPtr(
						"Common placeholder values in committed .env files are easy to miss.",
					),
					Severity:   "warning",
					SourceKind: "heuristic",
					SuppressionKey: model.StringPtr(
						fmt.Sprintf(
							"placeholder:%s:%d:%s",
							definition.FilePath,
							definition.LineNumber,
							definition.Name,
						),
					),
					Title:        "Placeholder-like committed value",
					VariableName: model.StringPtr(definition.Name),
				})
			}
		}
	}

	return order.Findings(findings)
}
