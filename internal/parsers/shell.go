package parsers

import (
	"regexp"

	"github.com/davisbuilds/envdiff/internal/lines"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

// Local assignment at the start of a line: `NAME=...` or `export NAME=...`.
var shellAssignPattern = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)=`)

func ScanShellFile(path string) (model.UsageScanResult, error) {
	fileLines, err := lines.Read(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	localDefs := shellLocalDefinitions(fileLines)

	usages := []model.EnvVarUsage{}
	for index, line := range fileLines {
		lineNumber := index + 1

		for _, match := range dollarBracedPattern.FindAllStringSubmatch(line, -1) {
			name := match[1]
			if _, defined := localDefs[name]; defined {
				continue
			}
			requiredness, defaultValue := expansionRequiredness(match[2], match[3])
			usages = append(usages, shellUsage(name, requiredness, defaultValue, path, lineNumber))
		}

		for _, match := range dollarSimplePattern.FindAllStringSubmatch(line, -1) {
			name := match[1]
			if _, defined := localDefs[name]; defined {
				continue
			}
			usages = append(usages, shellUsage(name, "required", nil, path, lineNumber))
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

func shellLocalDefinitions(fileLines []string) map[string]struct{} {
	defs := map[string]struct{}{}
	for _, line := range fileLines {
		if match := shellAssignPattern.FindStringSubmatch(line); match != nil {
			defs[match[1]] = struct{}{}
		}
	}
	return defs
}

func shellUsage(name string, requiredness string, defaultValue *string, path string, lineNumber int) model.EnvVarUsage {
	return model.EnvVarUsage{
		DefaultValue: defaultValue,
		FilePath:     path,
		LineNumber:   model.IntPtr(lineNumber),
		Name:         name,
		Requiredness: requiredness,
		SourceType:   "shell",
		UsageKind:    "shell_var",
	}
}
