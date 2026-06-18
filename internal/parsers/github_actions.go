package parsers

import (
	"regexp"
	"strings"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

var (
	actionsExpressionPattern = regexp.MustCompile(`\$\{\{\s*(.*?)\s*\}\}`)
	actionsReferencePattern  = regexp.MustCompile(`\b(secrets|vars)\.([A-Z_][A-Z0-9_]*)\b`)
)

func ScanGitHubActionsFile(path string) (model.UsageScanResult, error) {
	lines, err := readLines(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	usages := []model.EnvVarUsage{}
	for index, line := range lines {
		lineNumber := index + 1
		for _, expressionMatch := range actionsExpressionPattern.FindAllStringSubmatch(line, -1) {
			expression := expressionMatch[1]
			defaultValue := extractActionsDefault(expression)
			requiredness := "required"
			if defaultValue != nil {
				requiredness = "optional_with_default"
			}

			for _, reference := range actionsReferencePattern.FindAllStringSubmatch(expression, -1) {
				usages = append(usages, model.EnvVarUsage{
					DefaultValue: defaultValue,
					FilePath:     path,
					LineNumber:   model.IntPtr(lineNumber),
					Name:         reference[2],
					Requiredness: requiredness,
					SourceType:   "github_actions",
					UsageKind:    "github_actions_" + reference[1],
				})
			}
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

func extractActionsDefault(expression string) *string {
	_, fallback, ok := strings.Cut(expression, "||")
	if !ok {
		return nil
	}

	value := strings.TrimSpace(fallback)
	if value == "" {
		return nil
	}
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if first == last && (first == '"' || first == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return &value
}
