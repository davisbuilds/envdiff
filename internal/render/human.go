package render

import (
	"fmt"
	"strings"
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

func joinStrings(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
