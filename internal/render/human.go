package render

import (
	"fmt"
	"strings"

	"github.com/davisbuilds/envdiff/internal/model"
)

func CompareResult(result map[string]any) string {
	lines := []string{
		fmt.Sprintf("Compare %s vs %s", result["left_path"], result["right_path"]),
		fmt.Sprintf("Missing in left: %s", joinStrings(result["missing_in_left"].([]string))),
		fmt.Sprintf("Missing in right: %s", joinStrings(result["missing_in_right"].([]string))),
		fmt.Sprintf("Duplicates in left: %s", joinStrings(result["duplicates_in_left"].([]string))),
		fmt.Sprintf("Duplicates in right: %s", joinStrings(result["duplicates_in_right"].([]string))),
	}

	differing := result["differing_values"].([]map[string]any)
	if len(differing) > 0 {
		lines = append(lines, "Differing values:")
		for _, item := range differing {
			lines = append(
				lines,
				fmt.Sprintf(
					"  %s: %s=%s vs %s=%s",
					item["name"],
					item["left_kind"],
					item["left_value"],
					item["right_kind"],
					item["right_value"],
				),
			)
		}
	}

	return strings.Join(lines, "\n")
}

func MatrixResult(result map[string]any) string {
	lines := []string{
		fmt.Sprintf("Matrix files: %d", result["file_count"]),
		fmt.Sprintf("Variables: %d", result["variable_count"]),
		fmt.Sprintf("Inconsistent: %d", result["inconsistent_variable_count"]),
	}

	warnings := result["warnings"].([]string)
	if len(warnings) > 0 {
		lines = append(lines, fmt.Sprintf("Warnings: %d", len(warnings)))
	}

	variables := result["variables"].([]map[string]any)
	if len(variables) == 0 {
		lines = append(lines, "No variables matched the selected view.")
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "Variables:")
	for _, variable := range variables {
		reasons := []string{}
		if len(variable["missing_in"].([]string)) > 0 {
			reasons = append(reasons, "missing")
		}
		if len(variable["value_kinds"].([]string)) > 1 {
			reasons = append(reasons, "kind-mismatch")
		}
		if len(variable["duplicates_in"].([]string)) > 0 {
			reasons = append(reasons, "duplicate")
		}
		reasonText := strings.Join(reasons, ", ")
		if reasonText == "" {
			reasonText = "consistent"
		}
		lines = append(lines, fmt.Sprintf("  %s [%s]", variable["name"], reasonText))
		for _, fileEntry := range variable["files"].([]map[string]any) {
			if fileEntry["presence"] == "missing" {
				lines = append(lines, fmt.Sprintf("    %s: missing", fileEntry["path"]))
				continue
			}
			suffix := fileEntry["value_kind"].(string)
			if fileEntry["is_duplicate"].(bool) {
				suffix += ", duplicate"
			}
			lines = append(lines, fmt.Sprintf("    %s: present (%s)", fileEntry["path"], suffix))
		}
	}

	return strings.Join(lines, "\n")
}

func ScanResult(scanResult model.RepoScanResult) string {
	lines := []string{
		fmt.Sprintf("Scan root: %s", scanResult.RootPath),
		fmt.Sprintf("Definitions: %d", len(scanResult.Definitions)),
		fmt.Sprintf("Usages: %d", len(scanResult.Usages)),
		fmt.Sprintf("Contracts: %d", len(scanResult.Contracts)),
	}

	if len(scanResult.Contracts) > 0 {
		lines = append(lines, "Contracts:")
		for _, contract := range scanResult.Contracts {
			statuses := strings.Join(contract.Status, ",")
			if statuses == "" {
				statuses = "none"
			}
			lines = append(
				lines,
				fmt.Sprintf("  %s [%s] (%s)", contract.Name, contract.Requiredness, statuses),
			)
		}
	}

	return strings.Join(lines, "\n")
}

func GenerateResult(
	variableCount int,
	outputPath *string,
	annotate bool,
	checkPath *string,
	checkMatches *bool,
) string {
	if checkPath != nil && checkMatches != nil {
		if *checkMatches {
			return fmt.Sprintf("Generated output matches %s", *checkPath)
		}
		return fmt.Sprintf("Generated output drifted from %s", *checkPath)
	}

	suffix := ""
	if annotate {
		suffix = " with annotations"
	}
	if outputPath == nil {
		return fmt.Sprintf("Generated %d variables%s", variableCount, suffix)
	}
	return fmt.Sprintf("Generated %d variables%s to %s", variableCount, suffix, *outputPath)
}

func DoctorResult(
	rootPath string,
	findings []model.Finding,
	suppressedCount int,
	baselineWritten *string,
) string {
	counts := map[string]int{}
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	lines := []string{
		fmt.Sprintf("Doctor root: %s", rootPath),
		fmt.Sprintf("Findings: %d", len(findings)),
		fmt.Sprintf(
			"Summary: %d error, %d warning, %d info",
			counts["error"],
			counts["warning"],
			counts["info"],
		),
	}
	if suppressedCount > 0 {
		lines = append(lines, fmt.Sprintf("Suppressed: %d", suppressedCount))
	}
	if baselineWritten != nil {
		lines = append(lines, fmt.Sprintf("Baseline written: %s", *baselineWritten))
	}
	if len(findings) == 0 {
		lines = append(lines, "No active findings.")
		return strings.Join(lines, "\n")
	}

	for _, severity := range []string{"error", "warning", "info"} {
		scoped := []model.Finding{}
		for _, finding := range findings {
			if finding.Severity == severity {
				scoped = append(scoped, finding)
			}
		}
		if len(scoped) == 0 {
			continue
		}
		lines = append(lines, strings.ToUpper(severity[:1])+severity[1:]+"s:")
		for _, finding := range scoped {
			lines = append(lines, fmt.Sprintf("  %s %s", finding.Code, finding.Details))
		}
	}
	return strings.Join(lines, "\n")
}

func joinStrings(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
