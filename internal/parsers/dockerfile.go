package parsers

import (
	"regexp"

	"github.com/davisbuilds/envdiff/internal/lines"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

var (
	// ARG NAME[=default] — a single build-arg name.
	dockerArgPattern = regexp.MustCompile(`(?i)^\s*ARG\s+([A-Za-z_][A-Za-z0-9_]*)`)
	// ENV NAME=value (possibly several per line: ENV A=1 B=2).
	dockerEnvAssignPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)=`)
	// ENV NAME value (space form, single key).
	dockerEnvSpacePattern  = regexp.MustCompile(`(?i)^\s*ENV\s+([A-Za-z_][A-Za-z0-9_]*)\s+[^=]*$`)
	dockerEnvPrefixPattern = regexp.MustCompile(`(?i)^\s*ENV\s+`)
)

func ScanDockerfile(path string) (model.UsageScanResult, error) {
	fileLines, err := lines.Read(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	localDefs := dockerLocalDefinitions(fileLines)

	usages := []model.EnvVarUsage{}
	for index, line := range fileLines {
		lineNumber := index + 1

		for _, match := range dollarBracedPattern.FindAllStringSubmatch(line, -1) {
			name := match[1]
			if _, defined := localDefs[name]; defined {
				continue
			}
			requiredness, defaultValue := expansionRequiredness(match[2], match[3])
			usages = append(usages, dockerUsage(name, requiredness, defaultValue, path, lineNumber))
		}

		for _, match := range dollarSimplePattern.FindAllStringSubmatch(line, -1) {
			name := match[1]
			if _, defined := localDefs[name]; defined {
				continue
			}
			usages = append(usages, dockerUsage(name, "required", nil, path, lineNumber))
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

func dockerLocalDefinitions(fileLines []string) map[string]struct{} {
	defs := map[string]struct{}{}
	for _, line := range fileLines {
		if match := dockerArgPattern.FindStringSubmatch(line); match != nil {
			defs[match[1]] = struct{}{}
			continue
		}
		if !dockerEnvPrefixPattern.MatchString(line) {
			continue
		}
		if match := dockerEnvSpacePattern.FindStringSubmatch(line); match != nil {
			defs[match[1]] = struct{}{}
			continue
		}
		// ENV K1=v1 K2=v2 — capture every assigned key.
		for _, match := range dockerEnvAssignPattern.FindAllStringSubmatch(line, -1) {
			defs[match[1]] = struct{}{}
		}
	}
	return defs
}

func dockerUsage(name string, requiredness string, defaultValue *string, path string, lineNumber int) model.EnvVarUsage {
	return model.EnvVarUsage{
		DefaultValue: defaultValue,
		FilePath:     path,
		LineNumber:   model.IntPtr(lineNumber),
		Name:         name,
		Requiredness: requiredness,
		SourceType:   "dockerfile",
		UsageKind:    "dockerfile_var",
	}
}
