package analyzers

import (
	"github.com/davisbuilds/envdiff/internal/dotenv"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

func CompareDotenvFiles(left string, right string) (map[string]any, error) {
	leftResult, err := dotenv.Parse(left)
	if err != nil {
		return nil, err
	}
	rightResult, err := dotenv.Parse(right)
	if err != nil {
		return nil, err
	}

	leftLatest := latestByName(leftResult.Definitions)
	rightLatest := latestByName(rightResult.Definitions)
	leftNames := keys(leftLatest)
	rightNames := keys(rightLatest)
	sharedNames := intersection(leftNames, rightNames)

	differing := []map[string]any{}
	for _, name := range sharedNames {
		leftDefinition := leftLatest[name]
		rightDefinition := rightLatest[name]
		if leftDefinition.NormalizedValueKind != rightDefinition.NormalizedValueKind ||
			leftDefinition.Value != rightDefinition.Value {
			differing = append(differing, map[string]any{
				"left_kind":   leftDefinition.NormalizedValueKind,
				"left_value":  leftDefinition.Value,
				"name":        name,
				"right_kind":  rightDefinition.NormalizedValueKind,
				"right_value": rightDefinition.Value,
			})
		}
	}

	warnings := append([]string{}, leftResult.Warnings...)
	warnings = append(warnings, rightResult.Warnings...)

	return map[string]any{
		"differing_values":    differing,
		"duplicates_in_left":  duplicateNames(leftResult.Definitions),
		"duplicates_in_right": duplicateNames(rightResult.Definitions),
		"left_path":           left,
		"missing_in_left":     difference(rightNames, leftNames),
		"missing_in_right":    difference(leftNames, rightNames),
		"right_path":          right,
		"warnings":            order.Strings(warnings),
	}, nil
}

func latestByName(definitions []model.EnvVarDefinition) map[string]model.EnvVarDefinition {
	latest := map[string]model.EnvVarDefinition{}
	for _, definition := range definitions {
		latest[definition.Name] = definition
	}
	return latest
}

func duplicateNames(definitions []model.EnvVarDefinition) []string {
	names := []string{}
	for _, definition := range definitions {
		if definition.IsDuplicate {
			names = append(names, definition.Name)
		}
	}
	return order.Strings(names)
}

func keys(values map[string]model.EnvVarDefinition) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	return order.Strings(names)
}

func intersection(left []string, right []string) []string {
	rightSet := stringSet(right)
	values := []string{}
	for _, value := range left {
		if _, ok := rightSet[value]; ok {
			values = append(values, value)
		}
	}
	return order.Strings(values)
}

func difference(left []string, right []string) []string {
	rightSet := stringSet(right)
	values := []string{}
	for _, value := range left {
		if _, ok := rightSet[value]; !ok {
			values = append(values, value)
		}
	}
	return order.Strings(values)
}

func stringSet(values []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
