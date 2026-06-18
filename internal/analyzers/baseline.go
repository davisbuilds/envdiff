package analyzers

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
)

func BuildBaselineSnapshot(findings []model.Finding) model.BaselineSnapshot {
	entries := []model.BaselineEntry{}
	for _, finding := range findings {
		if finding.SuppressionKey == nil {
			continue
		}
		entries = append(entries, model.BaselineEntry{
			Code:           finding.Code,
			Reason:         finding.Reason,
			Severity:       finding.Severity,
			SuppressionKey: *finding.SuppressionKey,
			Title:          finding.Title,
			VariableName:   finding.VariableName,
		})
	}
	sort.SliceStable(entries, func(left int, right int) bool {
		return entries[left].SuppressionKey < entries[right].SuppressionKey
	})
	return model.BaselineSnapshot{Entries: entries, SchemaVersion: model.SchemaVersion}
}

func WriteBaselineSnapshot(path string, findings []model.Finding) (model.BaselineSnapshot, error) {
	snapshot := BuildBaselineSnapshot(findings)
	payload, err := render.MarshalIndent(snapshot)
	if err != nil {
		return model.BaselineSnapshot{}, err
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		return model.BaselineSnapshot{}, err
	}
	return snapshot, nil
}

func LoadBaselineSnapshot(path string) (model.BaselineSnapshot, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return model.BaselineSnapshot{}, err
	}
	var snapshot model.BaselineSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return model.BaselineSnapshot{}, err
	}
	return snapshot, nil
}

func LoadIgnoreKeys(path string) (map[string]struct{}, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	keys := map[string]struct{}{}
	for _, rawLine := range strings.Split(string(payload), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keys[line] = struct{}{}
	}
	return keys, nil
}

func ApplySuppressions(
	findings []model.Finding,
	suppressionKeys map[string]struct{},
) ([]model.Finding, []model.Finding) {
	active := []model.Finding{}
	suppressed := []model.Finding{}
	for _, finding := range findings {
		if finding.SuppressionKey != nil {
			if _, ok := suppressionKeys[*finding.SuppressionKey]; ok {
				suppressed = append(suppressed, finding)
				continue
			}
		}
		active = append(active, finding)
	}
	return active, suppressed
}
