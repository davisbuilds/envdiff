package model

import (
	"encoding/json"

	"github.com/davisbuilds/envdiff/internal/version"
)

type CommandMeta struct {
	Command       string `json:"command"`
	SchemaVersion string `json:"schema_version"`
}

type Location struct {
	ColumnNumber *int   `json:"column_number"`
	FilePath     string `json:"file_path"`
	LineNumber   *int   `json:"line_number"`
}

type EnvVarDefinition struct {
	FilePath            string   `json:"file_path"`
	IsDuplicate         bool     `json:"is_duplicate"`
	LineNumber          int      `json:"line_number"`
	Name                string   `json:"name"`
	NormalizedValueKind string   `json:"normalized_value_kind"`
	ParseWarnings       []string `json:"parse_warnings"`
	SourceType          string   `json:"source_type"`
	Value               string   `json:"value"`
}

type EnvVarUsage struct {
	DefaultValue *string `json:"default_value"`
	FilePath     string  `json:"file_path"`
	LineNumber   *int    `json:"line_number"`
	Name         string  `json:"name"`
	Requiredness string  `json:"requiredness"`
	SourceType   string  `json:"source_type"`
	UsageKind    string  `json:"usage_kind"`
}

type EnvVarContract struct {
	Aliases         []string           `json:"aliases"`
	Definitions     []EnvVarDefinition `json:"definitions"`
	Name            string             `json:"name"`
	PlaceholderLike bool               `json:"placeholder_like"`
	Requiredness    string             `json:"requiredness"`
	ResolutionNotes []string           `json:"resolution_notes"`
	SecretLike      bool               `json:"secret_like"`
	Status          []string           `json:"status"`
	Usages          []EnvVarUsage      `json:"usages"`
}

type DotenvParseResult struct {
	Definitions []EnvVarDefinition `json:"definitions"`
	Warnings    []string           `json:"warnings"`
}

type UsageScanResult struct {
	Usages   []EnvVarUsage `json:"usages"`
	Warnings []string      `json:"warnings"`
}

type ResolutionDecision struct {
	EnvFile     *string  `json:"env_file"`
	ExampleFile *string  `json:"example_file"`
	Notes       []string `json:"notes"`
	SourceFile  string   `json:"source_file"`
}

type RepoScanResult struct {
	Contracts   []EnvVarContract     `json:"contracts"`
	Definitions []EnvVarDefinition   `json:"definitions"`
	Resolutions []ResolutionDecision `json:"resolutions"`
	RootPath    string               `json:"root_path"`
	Usages      []EnvVarUsage        `json:"usages"`
	Warnings    []string             `json:"warnings"`
}

type BaselineEntry struct {
	Code           string  `json:"code"`
	Reason         *string `json:"reason"`
	Severity       string  `json:"severity"`
	SuppressionKey string  `json:"suppression_key"`
	Title          string  `json:"title"`
	VariableName   *string `json:"variable_name"`
}

type BaselineSnapshot struct {
	Entries       []BaselineEntry `json:"entries"`
	SchemaVersion string          `json:"schema_version"`
}

type Finding struct {
	Code             string     `json:"code"`
	Confidence       *string    `json:"confidence"`
	Details          string     `json:"details"`
	Locations        []Location `json:"locations"`
	Reason           *string    `json:"reason"`
	RelatedVariables []string   `json:"related_variables"`
	Severity         string     `json:"severity"`
	SourceKind       string     `json:"source_kind"`
	SuggestedFix     *string    `json:"suggested_fix"`
	SuppressionKey   *string    `json:"suppression_key"`
	Title            string     `json:"title"`
	VariableName     *string    `json:"variable_name"`
}

type SummaryCounts struct {
	Error   int `json:"error"`
	Info    int `json:"info"`
	Warning int `json:"warning"`
}

type JsonEnvelope struct {
	Data     map[string]any `json:"data"`
	Findings []Finding      `json:"findings"`
	Inputs   map[string]any `json:"inputs"`
	Meta     CommandMeta    `json:"meta"`
	Summary  SummaryCounts  `json:"summary"`
}

const SchemaVersion = version.SchemaVersion

func NewCommandMeta(command string) CommandMeta {
	return CommandMeta{Command: command, SchemaVersion: SchemaVersion}
}

func NewJsonEnvelope(command string, inputs map[string]any, data map[string]any) JsonEnvelope {
	return JsonEnvelope{
		Data:     data,
		Findings: []Finding{},
		Inputs:   inputs,
		Meta:     NewCommandMeta(command),
		Summary:  SummaryCounts{},
	}
}

func IntPtr(value int) *int {
	return &value
}

func StringPtr(value string) *string {
	return &value
}

func (meta CommandMeta) MarshalJSON() ([]byte, error) {
	type commandMeta CommandMeta
	if meta.SchemaVersion == "" {
		meta.SchemaVersion = SchemaVersion
	}
	return json.Marshal(commandMeta(meta))
}

func (definition EnvVarDefinition) MarshalJSON() ([]byte, error) {
	type envVarDefinition EnvVarDefinition
	definition.ParseWarnings = emptyStrings(definition.ParseWarnings)
	return json.Marshal(envVarDefinition(definition))
}

func (contract EnvVarContract) MarshalJSON() ([]byte, error) {
	type envVarContract EnvVarContract
	contract.Aliases = emptyStrings(contract.Aliases)
	contract.Definitions = emptyDefinitions(contract.Definitions)
	contract.ResolutionNotes = emptyStrings(contract.ResolutionNotes)
	contract.Status = emptyStrings(contract.Status)
	contract.Usages = emptyUsages(contract.Usages)
	if contract.Requiredness == "" {
		contract.Requiredness = "unknown"
	}
	return json.Marshal(envVarContract(contract))
}

func (result DotenvParseResult) MarshalJSON() ([]byte, error) {
	type dotenvParseResult DotenvParseResult
	result.Definitions = emptyDefinitions(result.Definitions)
	result.Warnings = emptyStrings(result.Warnings)
	return json.Marshal(dotenvParseResult(result))
}

func (result UsageScanResult) MarshalJSON() ([]byte, error) {
	type usageScanResult UsageScanResult
	result.Usages = emptyUsages(result.Usages)
	result.Warnings = emptyStrings(result.Warnings)
	return json.Marshal(usageScanResult(result))
}

func (decision ResolutionDecision) MarshalJSON() ([]byte, error) {
	type resolutionDecision ResolutionDecision
	decision.Notes = emptyStrings(decision.Notes)
	return json.Marshal(resolutionDecision(decision))
}

func (result RepoScanResult) MarshalJSON() ([]byte, error) {
	type repoScanResult RepoScanResult
	result.Contracts = emptyContracts(result.Contracts)
	result.Definitions = emptyDefinitions(result.Definitions)
	result.Resolutions = emptyResolutions(result.Resolutions)
	result.Usages = emptyUsages(result.Usages)
	result.Warnings = emptyStrings(result.Warnings)
	return json.Marshal(repoScanResult(result))
}

func (snapshot BaselineSnapshot) MarshalJSON() ([]byte, error) {
	type baselineSnapshot BaselineSnapshot
	snapshot.Entries = emptyBaselineEntries(snapshot.Entries)
	if snapshot.SchemaVersion == "" {
		snapshot.SchemaVersion = SchemaVersion
	}
	return json.Marshal(baselineSnapshot(snapshot))
}

func (finding Finding) MarshalJSON() ([]byte, error) {
	type findingJSON Finding
	finding.Locations = emptyLocations(finding.Locations)
	finding.RelatedVariables = emptyStrings(finding.RelatedVariables)
	if finding.SourceKind == "" {
		finding.SourceKind = "deterministic"
	}
	return json.Marshal(findingJSON(finding))
}

func (envelope JsonEnvelope) MarshalJSON() ([]byte, error) {
	type jsonEnvelope JsonEnvelope
	if envelope.Data == nil {
		envelope.Data = map[string]any{}
	}
	envelope.Findings = emptyFindings(envelope.Findings)
	if envelope.Inputs == nil {
		envelope.Inputs = map[string]any{}
	}
	if envelope.Meta.SchemaVersion == "" {
		envelope.Meta.SchemaVersion = SchemaVersion
	}
	return json.Marshal(jsonEnvelope(envelope))
}

func emptyStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func emptyDefinitions(values []EnvVarDefinition) []EnvVarDefinition {
	if values == nil {
		return []EnvVarDefinition{}
	}
	return values
}

func emptyUsages(values []EnvVarUsage) []EnvVarUsage {
	if values == nil {
		return []EnvVarUsage{}
	}
	return values
}

func emptyContracts(values []EnvVarContract) []EnvVarContract {
	if values == nil {
		return []EnvVarContract{}
	}
	return values
}

func emptyResolutions(values []ResolutionDecision) []ResolutionDecision {
	if values == nil {
		return []ResolutionDecision{}
	}
	return values
}

func emptyBaselineEntries(values []BaselineEntry) []BaselineEntry {
	if values == nil {
		return []BaselineEntry{}
	}
	return values
}

func emptyLocations(values []Location) []Location {
	if values == nil {
		return []Location{}
	}
	return values
}

func emptyFindings(values []Finding) []Finding {
	if values == nil {
		return []Finding{}
	}
	return values
}
