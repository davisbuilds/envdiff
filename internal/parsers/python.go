package parsers

import (
	"bufio"
	"os"
	"regexp"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

var (
	pythonEnvironPattern = regexp.MustCompile(`os\.environ\s*\[\s*["']([^"']+)["']\s*\]`)
	pythonGetenvPattern  = regexp.MustCompile(`os\.getenv\s*\(\s*["']([^"']+)["']\s*(,\s*["']([^"']*)["']\s*)?\)`)
)

func ScanPythonFile(path string) (model.UsageScanResult, error) {
	lines, err := readLines(path)
	if err != nil {
		return model.UsageScanResult{}, err
	}

	usages := []model.EnvVarUsage{}
	for index, line := range lines {
		lineNumber := index + 1
		for _, match := range pythonEnvironPattern.FindAllStringSubmatch(line, -1) {
			usages = append(usages, model.EnvVarUsage{
				DefaultValue: nil,
				FilePath:     path,
				LineNumber:   model.IntPtr(lineNumber),
				Name:         match[1],
				Requiredness: "required",
				SourceType:   "python",
				UsageKind:    "os.environ",
			})
		}

		matches := pythonGetenvPattern.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			name := line[match[2]:match[3]]
			var defaultValue *string
			requiredness := "optional"
			if match[4] != -1 {
				value := line[match[6]:match[7]]
				defaultValue = &value
				requiredness = "optional_with_default"
			}
			usages = append(usages, model.EnvVarUsage{
				DefaultValue: defaultValue,
				FilePath:     path,
				LineNumber:   model.IntPtr(lineNumber),
				Name:         name,
				Requiredness: requiredness,
				SourceType:   "python",
				UsageKind:    "os.getenv",
			})
		}
	}

	return model.UsageScanResult{Usages: order.Usages(usages), Warnings: []string{}}, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
