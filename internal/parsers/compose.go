package parsers

import (
	"regexp"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

var composeInterpolationPattern = regexp.MustCompile(`\$\{([A-Z0-9_]+)(:-(.*?))?\}`)

func ScanComposeFile(path string) (model.UsageScanResult, error) {
	lines, err := readLines(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	usages := []model.EnvVarUsage{}
	for index, line := range lines {
		lineNumber := index + 1
		for _, match := range composeInterpolationPattern.FindAllStringSubmatchIndex(line, -1) {
			name := line[match[2]:match[3]]
			var defaultValue *string
			requiredness := "required"
			if match[4] != -1 {
				value := line[match[6]:match[7]]
				defaultValue = &value
				requiredness = "optional_with_default"
			}
			usages = append(usages, model.EnvVarUsage{
				DefaultValue: defaultValue,
				FilePath:     path,
				LineNumber:   model.IntPtr(lineNumber),
				Name:         name,
				Requiredness: requiredness,
				SourceType:   "compose",
				UsageKind:    "compose_interpolation",
			})
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}
