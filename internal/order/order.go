package order

import (
	"sort"

	"github.com/davisbuilds/envdiff/internal/model"
)

func Strings(values []string) []string {
	ordered := append([]string(nil), values...)
	sort.Strings(ordered)
	return ordered
}

func Definitions(values []model.EnvVarDefinition) []model.EnvVarDefinition {
	ordered := append([]model.EnvVarDefinition(nil), values...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		leftValue := ordered[left]
		rightValue := ordered[right]
		if leftValue.Name != rightValue.Name {
			return leftValue.Name < rightValue.Name
		}
		if leftValue.FilePath != rightValue.FilePath {
			return leftValue.FilePath < rightValue.FilePath
		}
		return leftValue.LineNumber < rightValue.LineNumber
	})
	return ordered
}

func Usages(values []model.EnvVarUsage) []model.EnvVarUsage {
	ordered := append([]model.EnvVarUsage(nil), values...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		leftValue := ordered[left]
		rightValue := ordered[right]
		if leftValue.Name != rightValue.Name {
			return leftValue.Name < rightValue.Name
		}
		if leftValue.FilePath != rightValue.FilePath {
			return leftValue.FilePath < rightValue.FilePath
		}
		return lineNumberValue(leftValue.LineNumber) < lineNumberValue(rightValue.LineNumber)
	})
	return ordered
}

func Contracts(values []model.EnvVarContract) []model.EnvVarContract {
	ordered := append([]model.EnvVarContract(nil), values...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		return ordered[left].Name < ordered[right].Name
	})
	return ordered
}

func Findings(values []model.Finding) []model.Finding {
	ordered := append([]model.Finding(nil), values...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		leftValue := ordered[left]
		rightValue := ordered[right]
		if severityRank(leftValue.Severity) != severityRank(rightValue.Severity) {
			return severityRank(leftValue.Severity) < severityRank(rightValue.Severity)
		}
		if leftValue.Code != rightValue.Code {
			return leftValue.Code < rightValue.Code
		}
		if stringValue(leftValue.VariableName) != stringValue(rightValue.VariableName) {
			return stringValue(leftValue.VariableName) < stringValue(rightValue.VariableName)
		}
		return locationsLess(leftValue.Locations, rightValue.Locations)
	})
	return ordered
}

func severityRank(severity string) int {
	switch severity {
	case "error":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 99
	}
}

func lineNumberValue(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func locationsLess(left []model.Location, right []model.Location) bool {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for index := range limit {
		leftLocation := left[index]
		rightLocation := right[index]
		if leftLocation.FilePath != rightLocation.FilePath {
			return leftLocation.FilePath < rightLocation.FilePath
		}
		leftLine := lineNumberValue(leftLocation.LineNumber)
		rightLine := lineNumberValue(rightLocation.LineNumber)
		if leftLine != rightLine {
			return leftLine < rightLine
		}
	}
	return len(left) < len(right)
}
