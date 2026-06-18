package analyzers

import (
	"github.com/davisbuilds/envdiff/internal/dotenv"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

func MatrixDotenvFiles(paths []string, showAll bool) (map[string]any, error) {
	type parsedFile struct {
		path   string
		result model.DotenvParseResult
	}

	parsed := []parsedFile{}
	latestByPath := map[string]map[string]model.EnvVarDefinition{}
	definitionsByPath := map[string][]model.EnvVarDefinition{}
	allNameSet := map[string]struct{}{}
	for _, path := range paths {
		result, err := dotenv.Parse(path)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, parsedFile{path: path, result: result})
		latest := latestByName(result.Definitions)
		latestByPath[path] = latest
		definitionsByPath[path] = result.Definitions
		for name := range latest {
			allNameSet[name] = struct{}{}
		}
	}

	allNames := make([]string, 0, len(allNameSet))
	for name := range allNameSet {
		allNames = append(allNames, name)
	}
	allNames = order.Strings(allNames)

	variables := []map[string]any{}
	inconsistentCount := 0
	for _, name := range allNames {
		files := []map[string]any{}
		kindSet := map[string]struct{}{}
		presentIn := []string{}
		missingIn := []string{}
		duplicatesIn := []string{}

		for _, path := range paths {
			definition, ok := latestByPath[path][name]
			if !ok {
				files = append(files, map[string]any{"path": path, "presence": "missing"})
				missingIn = append(missingIn, path)
				continue
			}

			kindSet[definition.NormalizedValueKind] = struct{}{}
			presentIn = append(presentIn, path)
			isDuplicate := hasDuplicateDefinition(definitionsByPath[path], name)
			if isDuplicate {
				duplicatesIn = append(duplicatesIn, path)
			}
			files = append(files, map[string]any{
				"is_duplicate": isDuplicate,
				"path":         path,
				"presence":     "present",
				"value_kind":   definition.NormalizedValueKind,
			})
		}

		valueKinds := make([]string, 0, len(kindSet))
		for kind := range kindSet {
			valueKinds = append(valueKinds, kind)
		}
		valueKinds = order.Strings(valueKinds)
		inconsistent := len(missingIn) > 0 || len(valueKinds) > 1 || len(duplicatesIn) > 0
		if inconsistent {
			inconsistentCount++
		}

		variable := map[string]any{
			"duplicates_in": duplicatesIn,
			"files":         files,
			"missing_in":    missingIn,
			"name":          name,
			"present_in":    presentIn,
			"status":        "consistent",
			"value_kinds":   valueKinds,
		}
		if inconsistent {
			variable["status"] = "inconsistent"
		}
		if showAll || inconsistent {
			variables = append(variables, variable)
		}
	}

	warnings := []string{}
	for _, item := range parsed {
		warnings = append(warnings, item.result.Warnings...)
	}

	return map[string]any{
		"file_count":                  len(paths),
		"inconsistent_variable_count": inconsistentCount,
		"paths":                       append([]string(nil), paths...),
		"show_all":                    showAll,
		"variable_count":              len(allNames),
		"variables":                   variables,
		"warnings":                    order.Strings(warnings),
	}, nil
}

func hasDuplicateDefinition(definitions []model.EnvVarDefinition, name string) bool {
	for _, definition := range definitions {
		if definition.Name == name && definition.IsDuplicate {
			return true
		}
	}
	return false
}
