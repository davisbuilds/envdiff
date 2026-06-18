package analyzers

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
)

func DoctorRepository(scanResult model.RepoScanResult) []model.Finding {
	definitionsByFile := map[string][]model.EnvVarDefinition{}
	resolutionsBySource := map[string]model.ResolutionDecision{}
	associatedUsageNames := map[string]map[string]struct{}{}
	findings := []model.Finding{}
	seen := map[string]struct{}{}

	for _, resolution := range scanResult.Resolutions {
		resolutionsBySource[resolution.SourceFile] = resolution
	}

	for _, definition := range scanResult.Definitions {
		definitionsByFile[definition.FilePath] = append(
			definitionsByFile[definition.FilePath],
			definition,
		)
		if definition.IsDuplicate {
			findings = append(findings, model.Finding{
				Code: "ENV006",
				Details: fmt.Sprintf(
					"%s is defined more than once in %s.",
					definition.Name,
					definition.FilePath,
				),
				Locations: []model.Location{{
					FilePath:   definition.FilePath,
					LineNumber: model.IntPtr(definition.LineNumber),
				}},
				Reason:   model.StringPtr("Duplicate keys create ambiguous effective values."),
				Severity: "warning",
				SuppressionKey: model.StringPtr(
					fmt.Sprintf("duplicate:%s:%s", definition.FilePath, definition.Name),
				),
				Title:        "Duplicate definition",
				VariableName: model.StringPtr(definition.Name),
			})
		}
	}

	for _, usage := range scanResult.Usages {
		resolution, hasResolution := resolutionsBySource[usage.FilePath]
		envNames := definitionNames(definitionsByFile, nil)
		exampleNames := definitionNames(definitionsByFile, nil)
		if hasResolution {
			envNames = definitionNames(definitionsByFile, resolution.EnvFile)
			exampleNames = definitionNames(definitionsByFile, resolution.ExampleFile)
			if resolution.EnvFile != nil {
				addAssociatedName(associatedUsageNames, *resolution.EnvFile, usage.Name)
			}
			if resolution.ExampleFile != nil {
				addAssociatedName(associatedUsageNames, *resolution.ExampleFile, usage.Name)
			}
		}

		if _, ok := envNames[usage.Name]; !ok {
			code := "ENV002"
			severity := "warning"
			if usage.Requiredness == "required" {
				code = "ENV001"
				severity = "error"
			}
			key := fmt.Sprintf("%s:%s:%s", code, usage.Name, usage.FilePath)
			if _, alreadySeen := seen[key]; !alreadySeen {
				seen[key] = struct{}{}
				findings = append(findings, model.Finding{
					Code:    code,
					Details: missingDetails(usage, resolution, hasResolution),
					Locations: []model.Location{{
						FilePath:   usage.FilePath,
						LineNumber: usage.LineNumber,
					}},
					Reason: model.StringPtr(
						"The associated .env file does not define a referenced variable.",
					),
					Severity: severity,
					SuppressionKey: model.StringPtr(
						fmt.Sprintf("missing:%s:%s", usage.FilePath, usage.Name),
					),
					Title:        "Missing variable",
					VariableName: model.StringPtr(usage.Name),
				})
			}
			findings = append(
				findings,
				aliasFindingsForMissingUsage(usage, envNames, seen)...,
			)
		}

		if hasResolution && resolution.ExampleFile != nil {
			if _, ok := exampleNames[usage.Name]; !ok {
				key := fmt.Sprintf("ENV004:%s:%s", usage.Name, *resolution.ExampleFile)
				if _, alreadySeen := seen[key]; !alreadySeen {
					seen[key] = struct{}{}
					findings = append(findings, model.Finding{
						Code: "ENV004",
						Details: fmt.Sprintf(
							"%s is referenced by %s but absent from %s.",
							usage.Name,
							usage.FilePath,
							*resolution.ExampleFile,
						),
						Locations: []model.Location{{
							FilePath:   usage.FilePath,
							LineNumber: usage.LineNumber,
						}},
						Reason: model.StringPtr(
							"Referenced variables should appear in the nearest .env.example.",
						),
						Severity: "warning",
						SuppressionKey: model.StringPtr(
							fmt.Sprintf("undocumented:%s:%s", *resolution.ExampleFile, usage.Name),
						),
						Title:        "Undocumented variable",
						VariableName: model.StringPtr(usage.Name),
					})
				}
			}
		}
	}

	for _, definition := range scanResult.Definitions {
		if filepath.Base(definition.FilePath) == ".env.example" {
			continue
		}
		if _, ok := associatedUsageNames[definition.FilePath][definition.Name]; ok {
			continue
		}
		key := fmt.Sprintf("ENV003:%s:%s", definition.Name, definition.FilePath)
		if _, alreadySeen := seen[key]; alreadySeen {
			continue
		}
		seen[key] = struct{}{}
		findings = append(findings, model.Finding{
			Code: "ENV003",
			Details: fmt.Sprintf(
				"%s is defined in %s but not referenced.",
				definition.Name,
				definition.FilePath,
			),
			Locations: []model.Location{{
				FilePath:   definition.FilePath,
				LineNumber: model.IntPtr(definition.LineNumber),
			}},
			Reason:   model.StringPtr("Defined variables without matching usage may be stale."),
			Severity: "info",
			SuppressionKey: model.StringPtr(
				fmt.Sprintf("unused:%s:%s", definition.FilePath, definition.Name),
			),
			Title:        "Unused variable",
			VariableName: model.StringPtr(definition.Name),
		})
	}

	findings = append(findings, skewFindings(scanResult, definitionsByFile, associatedUsageNames, seen)...)
	findings = append(findings, SecretAndPlaceholderFindings(scanResult)...)
	return order.Findings(findings)
}

func SummarizeFindings(findings []model.Finding) model.SummaryCounts {
	summary := model.SummaryCounts{}
	for _, finding := range findings {
		switch finding.Severity {
		case "error":
			summary.Error++
		case "warning":
			summary.Warning++
		case "info":
			summary.Info++
		}
	}
	return summary
}

func ShouldFail(summary model.SummaryCounts, threshold string) (bool, error) {
	switch strings.ToLower(threshold) {
	case "error":
		return summary.Error > 0, nil
	case "warning":
		return summary.Error > 0 || summary.Warning > 0, nil
	case "info":
		return summary.Error > 0 || summary.Warning > 0 || summary.Info > 0, nil
	default:
		return false, fmt.Errorf("fail-on must be one of: error, warning, info")
	}
}

func definitionNames(
	definitionsByFile map[string][]model.EnvVarDefinition,
	filePath *string,
) map[string]struct{} {
	names := map[string]struct{}{}
	if filePath == nil {
		return names
	}
	for _, definition := range definitionsByFile[*filePath] {
		names[definition.Name] = struct{}{}
	}
	return names
}

func addAssociatedName(associated map[string]map[string]struct{}, filePath string, name string) {
	if associated[filePath] == nil {
		associated[filePath] = map[string]struct{}{}
	}
	associated[filePath][name] = struct{}{}
}

func missingDetails(
	usage model.EnvVarUsage,
	resolution model.ResolutionDecision,
	hasResolution bool,
) string {
	if hasResolution && resolution.EnvFile != nil {
		return fmt.Sprintf(
			"%s is referenced by %s but missing from %s.",
			usage.Name,
			usage.FilePath,
			*resolution.EnvFile,
		)
	}
	return fmt.Sprintf(
		"%s is referenced by %s but no associated .env defines it.",
		usage.Name,
		usage.FilePath,
	)
}

func aliasFindingsForMissingUsage(
	usage model.EnvVarUsage,
	envNames map[string]struct{},
	seen map[string]struct{},
) []model.Finding {
	findings := []model.Finding{}
	for _, candidate := range FindAliasCandidates(usage.Name, envNames) {
		key := fmt.Sprintf("ENV007:%s:%s", usage.Name, candidate.CandidateName)
		if _, alreadySeen := seen[key]; alreadySeen {
			continue
		}
		seen[key] = struct{}{}
		findings = append(findings, model.Finding{
			Code:       "ENV007",
			Confidence: model.StringPtr("low"),
			Details: fmt.Sprintf(
				"%s is missing, but nearby definitions include %s.",
				usage.Name,
				candidate.CandidateName,
			),
			Locations: []model.Location{{
				FilePath:   usage.FilePath,
				LineNumber: usage.LineNumber,
			}},
			Reason:           model.StringPtr(candidate.Reason),
			RelatedVariables: []string{candidate.CandidateName},
			Severity:         "warning",
			SourceKind:       "heuristic",
			SuppressionKey: model.StringPtr(
				fmt.Sprintf("alias:%s:%s", usage.Name, candidate.CandidateName),
			),
			Title:        "Possible alias candidate",
			VariableName: model.StringPtr(usage.Name),
		})
	}
	return findings
}

func skewFindings(
	scanResult model.RepoScanResult,
	definitionsByFile map[string][]model.EnvVarDefinition,
	associatedUsageNames map[string]map[string]struct{},
	seen map[string]struct{},
) []model.Finding {
	findings := []model.Finding{}
	for _, resolution := range scanResult.Resolutions {
		if resolution.EnvFile == nil || resolution.ExampleFile == nil {
			continue
		}
		envNames := definitionNames(definitionsByFile, resolution.EnvFile)
		exampleNames := definitionNames(definitionsByFile, resolution.ExampleFile)
		associatedNames := map[string]struct{}{}
		for name := range associatedUsageNames[*resolution.EnvFile] {
			associatedNames[name] = struct{}{}
		}
		for name := range associatedUsageNames[*resolution.ExampleFile] {
			associatedNames[name] = struct{}{}
		}

		names := make([]string, 0, len(exampleNames))
		for name := range exampleNames {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if _, ok := envNames[name]; ok {
				continue
			}
			if _, ok := associatedNames[name]; ok {
				continue
			}
			key := fmt.Sprintf("ENV005:%s:%s", name, *resolution.ExampleFile)
			if _, alreadySeen := seen[key]; alreadySeen {
				continue
			}
			seen[key] = struct{}{}
			findings = append(findings, model.Finding{
				Code: "ENV005",
				Details: fmt.Sprintf(
					"%s appears in %s but is absent from %s and not referenced.",
					name,
					*resolution.ExampleFile,
					*resolution.EnvFile,
				),
				Locations: []model.Location{{FilePath: *resolution.ExampleFile}},
				Reason: model.StringPtr(
					"Template-only variables with no matching local definition or usage are likely stale.",
				),
				Severity: "info",
				SuppressionKey: model.StringPtr(
					fmt.Sprintf("skew:%s:%s", *resolution.ExampleFile, name),
				),
				Title:        "Template skew",
				VariableName: model.StringPtr(name),
			})
		}
	}
	return findings
}
