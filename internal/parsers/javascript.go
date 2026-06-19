package parsers

import (
	"regexp"
	"strings"

	"github.com/davisbuilds/envdiff/internal/lines"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

// Env access in two literal forms (dot and bracket) for both process.env (Node)
// and import.meta.env (Vite), each with an optional inline fallback (|| or ??)
// capturing a quoted string, number, or boolean literal. Computed keys
// (process.env[expr]) and whole-object access are intentionally not matched.
var (
	jsDefault = `('[^']*'|"[^"]*"|\d[\w.]*|true|false)`

	jsProcessEnvDotPattern     = jsDotPattern(`process\.env`)
	jsProcessEnvBracketPattern = jsBracketPattern(`process\.env`)
	jsImportMetaDotPattern     = jsDotPattern(`import\.meta\.env`)
	jsImportMetaBracketPattern = jsBracketPattern(`import\.meta\.env`)

	// Single-line destructuring: `const { A, B: alias } = process.env`. The
	// captured object decides the usage_kind; each member's source key is taken.
	jsDestructurePattern  = regexp.MustCompile(`\{([^{}]+)\}\s*=\s*(process\.env|import\.meta\.env)\b`)
	jsDestructureMemberID = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)`)

	jsAccessPatterns = []struct {
		pattern   *regexp.Regexp
		usageKind string
	}{
		{jsProcessEnvDotPattern, "process.env"},
		{jsProcessEnvBracketPattern, "process.env"},
		{jsImportMetaDotPattern, "import.meta.env"},
		{jsImportMetaBracketPattern, "import.meta.env"},
	}
)

func jsDotPattern(prefix string) *regexp.Regexp {
	return regexp.MustCompile(
		prefix + `\.([A-Za-z_][A-Za-z0-9_]*)(?:\s*(?:\|\||\?\?)\s*` + jsDefault + `)?`,
	)
}

func jsBracketPattern(prefix string) *regexp.Regexp {
	return regexp.MustCompile(
		prefix + `\[\s*["']([A-Za-z_][A-Za-z0-9_]*)["']\s*\](?:\s*(?:\|\||\?\?)\s*` + jsDefault + `)?`,
	)
}

func ScanJavaScriptFile(path string) (model.UsageScanResult, error) {
	fileLines, err := lines.Read(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	usages := []model.EnvVarUsage{}
	for index, line := range fileLines {
		lineNumber := index + 1
		for _, access := range jsAccessPatterns {
			for _, match := range access.pattern.FindAllStringSubmatch(line, -1) {
				usages = append(usages, jsUsage(match[1], match[2], access.usageKind, path, lineNumber))
			}
		}
		for _, name := range jsDestructuredNames(line) {
			usages = append(usages, jsUsage(name.name, "", name.usageKind, path, lineNumber))
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

type jsDestructuredName struct {
	name      string
	usageKind string
}

// jsDestructuredNames extracts the source keys of a single-line destructuring of
// process.env / import.meta.env (e.g. `const { A, B: alias } = process.env`).
func jsDestructuredNames(line string) []jsDestructuredName {
	names := []jsDestructuredName{}
	for _, block := range jsDestructurePattern.FindAllStringSubmatch(line, -1) {
		usageKind := block[2]
		for _, member := range strings.Split(block[1], ",") {
			if id := jsDestructureMemberID.FindStringSubmatch(member); id != nil {
				names = append(names, jsDestructuredName{name: id[1], usageKind: usageKind})
			}
		}
	}
	return names
}

// jsUsage builds a usage from a captured name and (possibly empty) default
// literal. Bare reads are optional (JS returns undefined rather than throwing);
// an inline fallback makes them optional_with_default.
func jsUsage(name string, defaultLiteral string, usageKind string, path string, lineNumber int) model.EnvVarUsage {
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
		UsageKind:    usageKind,
	}
}
