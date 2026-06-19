package parsers

import (
	"regexp"

	"github.com/davisbuilds/envdiff/internal/lines"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

// process.env access in two literal forms, each with an optional inline
// fallback (|| or ??) capturing a quoted string, number, or boolean literal.
// Dynamic keys (process.env[expr]), whole-object access (process.env), and
// destructuring are intentionally not matched.
var (
	jsDefault              = `('[^']*'|"[^"]*"|\d[\w.]*|true|false)`
	jsProcessEnvDotPattern = regexp.MustCompile(
		`process\.env\.([A-Za-z_][A-Za-z0-9_]*)(?:\s*(?:\|\||\?\?)\s*` + jsDefault + `)?`,
	)
	jsProcessEnvBracketPattern = regexp.MustCompile(
		`process\.env\[\s*["']([A-Za-z_][A-Za-z0-9_]*)["']\s*\](?:\s*(?:\|\||\?\?)\s*` + jsDefault + `)?`,
	)
)

func ScanJavaScriptFile(path string) (model.UsageScanResult, error) {
	fileLines, err := lines.Read(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	usages := []model.EnvVarUsage{}
	for index, line := range fileLines {
		lineNumber := index + 1
		for _, pattern := range []*regexp.Regexp{jsProcessEnvDotPattern, jsProcessEnvBracketPattern} {
			for _, match := range pattern.FindAllStringSubmatch(line, -1) {
				usages = append(usages, jsUsage(match[1], match[2], path, lineNumber))
			}
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

// jsUsage builds a usage from a captured name and (possibly empty) default
// literal. Bare reads are optional (JS returns undefined rather than throwing);
// an inline fallback makes them optional_with_default.
func jsUsage(name string, defaultLiteral string, path string, lineNumber int) model.EnvVarUsage {
	requiredness := "optional"
	var defaultValue *string
	if defaultLiteral != "" {
		value := stripMatchingQuotes(defaultLiteral)
		defaultValue = &value
		requiredness = "optional_with_default"
	}
	return model.EnvVarUsage{
		DefaultValue: defaultValue,
		FilePath:     path,
		LineNumber:   model.IntPtr(lineNumber),
		Name:         name,
		Requiredness: requiredness,
		SourceType:   "javascript",
		UsageKind:    "process.env",
	}
}
