package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/normalize"
	"github.com/davisbuilds/envdiff/internal/order"
)

var keyValuePattern = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)

func Parse(path string) (model.DotenvParseResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return model.DotenvParseResult{}, err
	}
	defer file.Close()

	seen := map[string]int{}
	definitions := []model.EnvVarDefinition{}
	warnings := []string{}

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()
		stripped := strings.TrimSpace(rawLine)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}

		match := keyValuePattern.FindStringSubmatch(rawLine)
		if match == nil {
			warnings = append(
				warnings,
				fmt.Sprintf("%s:%d: unsupported dotenv syntax", path, lineNumber),
			)
			continue
		}

		name := match[1]
		value := parseValue(match[2])
		duplicateIndex := seen[name]
		seen[name] = duplicateIndex + 1

		definitions = append(definitions, model.EnvVarDefinition{
			FilePath:            path,
			IsDuplicate:         duplicateIndex > 0,
			LineNumber:          lineNumber,
			Name:                name,
			NormalizedValueKind: normalize.ValueKind(value),
			ParseWarnings:       []string{},
			SourceType:          "dotenv",
			Value:               value,
		})
	}
	if err := scanner.Err(); err != nil {
		return model.DotenvParseResult{}, err
	}

	return model.DotenvParseResult{
		Definitions: order.Definitions(definitions),
		Warnings:    order.Strings(warnings),
	}, nil
}

func parseValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if first == last && (first == '"' || first == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
