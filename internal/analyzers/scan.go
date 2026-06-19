package analyzers

import (
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/davisbuilds/envdiff/internal/dotenv"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/order"
	"github.com/davisbuilds/envdiff/internal/parsers"
	"github.com/davisbuilds/envdiff/internal/paths"
)

var composeFilenames = map[string]struct{}{
	"docker-compose.yml":  {},
	"docker-compose.yaml": {},
	"compose.yml":         {},
	"compose.yaml":        {},
}

var javaScriptExtensions = map[string]struct{}{
	".js":  {},
	".jsx": {},
	".ts":  {},
	".tsx": {},
	".mjs": {},
	".cjs": {},
}

type contractPayload struct {
	definitions []model.EnvVarDefinition
	usages      []model.EnvVarUsage
	notes       []string
}

// fileScan holds one file's parsed contribution to the repository result.
type fileScan struct {
	definitions []model.EnvVarDefinition
	usages      []model.EnvVarUsage
	warnings    []string
	resolution  *model.ResolutionDecision
	err         error
}

func ScanRepository(path string) (model.RepoScanResult, error) {
	root, err := paths.Canonical(path)
	if err != nil {
		return model.RepoScanResult{}, err
	}

	files, err := paths.IterRepoFiles(root, nil)
	if err != nil {
		return model.RepoScanResult{}, err
	}

	// Files are parsed concurrently (each parse reads and regex-scans its own
	// file). Results are written to a per-file slot and merged in file order,
	// so the output is independent of goroutine scheduling.
	results := scanFilesConcurrently(files, root)

	definitions := []model.EnvVarDefinition{}
	usages := []model.EnvVarUsage{}
	warnings := []string{}
	resolutionMap := map[string]model.ResolutionDecision{}

	for index, result := range results {
		if result.err != nil {
			return model.RepoScanResult{}, result.err
		}
		definitions = append(definitions, result.definitions...)
		usages = append(usages, result.usages...)
		warnings = append(warnings, result.warnings...)
		if result.resolution != nil {
			resolutionMap[files[index]] = *result.resolution
		}
	}

	resolutions := make([]model.ResolutionDecision, 0, len(resolutionMap))
	for _, resolution := range resolutionMap {
		resolutions = append(resolutions, resolution)
	}
	sort.SliceStable(resolutions, func(left int, right int) bool {
		return resolutions[left].SourceFile < resolutions[right].SourceFile
	})

	return model.RepoScanResult{
		Contracts:   buildContracts(definitions, usages, resolutionMap),
		Definitions: order.Definitions(definitions),
		Resolutions: resolutions,
		RootPath:    root,
		Usages:      order.Usages(usages),
		Warnings:    order.Strings(warnings),
	}, nil
}

// scanFilesConcurrently parses each file in a bounded worker pool and returns
// one result per file, in the same order as files.
func scanFilesConcurrently(files []string, root string) []fileScan {
	results := make([]fileScan, len(files))
	if len(files) == 0 {
		return results
	}

	workers := runtime.GOMAXPROCS(0)
	if workers > len(files) {
		workers = len(files)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each index is owned by a single goroutine, so writing distinct
			// slice elements concurrently needs no further synchronization.
			for index := range jobs {
				results[index] = scanFile(files[index], root)
			}
		}()
	}
	for index := range files {
		jobs <- index
	}
	close(jobs)
	wg.Wait()

	return results
}

// scanFile parses a single file according to its kind and returns its
// contribution to the repository result.
func scanFile(filePath string, root string) fileScan {
	name := filepath.Base(filePath)
	switch {
	case name == ".env" || name == ".env.example":
		result, err := dotenv.Parse(filePath)
		if err != nil {
			return fileScan{err: err}
		}
		return fileScan{definitions: result.Definitions, warnings: result.Warnings}
	case filepath.Ext(filePath) == ".py":
		result, err := parsers.ScanPythonFile(filePath)
		return usageFileScan(result, err, filePath, root)
	case isJavaScriptFile(filePath):
		result, err := parsers.ScanJavaScriptFile(filePath)
		return usageFileScan(result, err, filePath, root)
	case isComposeFile(name):
		result, err := parsers.ScanComposeFile(filePath)
		return usageFileScan(result, err, filePath, root)
	case isGitHubActionsWorkflow(filePath):
		result, err := parsers.ScanGitHubActionsFile(filePath)
		return usageFileScan(result, err, filePath, root)
	}
	return fileScan{}
}

// usageFileScan adapts a parser result into a fileScan, attaching the dotenv
// resolution decision for the source file.
func usageFileScan(
	result model.UsageScanResult,
	err error,
	filePath string,
	root string,
) fileScan {
	if err != nil {
		return fileScan{err: err}
	}
	resolution := resolveUsageFile(filePath, root)
	return fileScan{
		usages:     result.Usages,
		warnings:   result.Warnings,
		resolution: &resolution,
	}
}

func resolveUsageFile(filePath string, root string) model.ResolutionDecision {
	envFile, _ := paths.FindNearestNamedFile(filePath, root, ".env")
	exampleFile, _ := paths.FindNearestNamedFile(filePath, root, ".env.example")
	notes := []string{}
	if envFile != nil {
		notes = append(notes, "env:"+*envFile)
	}
	if exampleFile != nil {
		notes = append(notes, "example:"+*exampleFile)
	}
	if len(notes) == 0 {
		notes = append(notes, "no associated dotenv files found")
	}

	return model.ResolutionDecision{
		EnvFile:     envFile,
		ExampleFile: exampleFile,
		Notes:       notes,
		SourceFile:  filePath,
	}
}

func buildContracts(
	definitions []model.EnvVarDefinition,
	usages []model.EnvVarUsage,
	resolutionMap map[string]model.ResolutionDecision,
) []model.EnvVarContract {
	byName := map[string]*contractPayload{}

	for _, definition := range definitions {
		payload := payloadForName(byName, definition.Name)
		payload.definitions = append(payload.definitions, definition)
	}

	for _, usage := range usages {
		payload := payloadForName(byName, usage.Name)
		payload.usages = append(payload.usages, usage)
		if resolution, ok := resolutionMap[usage.FilePath]; ok {
			payload.notes = append(payload.notes, resolution.Notes...)
		}
	}

	contracts := []model.EnvVarContract{}
	for name, payload := range byName {
		statuses := []string{}
		if len(payload.usages) > 0 {
			statuses = append(statuses, "referenced")
		}
		if len(payload.definitions) > 0 {
			statuses = append(statuses, "defined")
		}
		if len(payload.usages) > 0 && len(payload.definitions) == 0 {
			statuses = append(statuses, "undefined")
		}
		if len(payload.definitions) > 0 && len(payload.usages) == 0 {
			statuses = append(statuses, "unreferenced")
		}

		contracts = append(contracts, model.EnvVarContract{
			Aliases:         []string{},
			Definitions:     order.Definitions(payload.definitions),
			Name:            name,
			PlaceholderLike: hasValueKind(payload.definitions, "placeholder"),
			Requiredness:    inferRequiredness(payload.usages),
			ResolutionNotes: uniqueSorted(payload.notes),
			SecretLike:      hasValueKind(payload.definitions, "secret_like"),
			Status:          order.Strings(statuses),
			Usages:          order.Usages(payload.usages),
		})
	}

	return order.Contracts(contracts)
}

func payloadForName(byName map[string]*contractPayload, name string) *contractPayload {
	payload, ok := byName[name]
	if !ok {
		payload = &contractPayload{}
		byName[name] = payload
	}
	return payload
}

func inferRequiredness(usages []model.EnvVarUsage) string {
	requirednesses := map[string]struct{}{}
	for _, usage := range usages {
		requirednesses[usage.Requiredness] = struct{}{}
	}
	if _, ok := requirednesses["required"]; ok {
		return "required"
	}
	if _, ok := requirednesses["optional_with_default"]; ok {
		return "optional_with_default"
	}
	if _, ok := requirednesses["optional"]; ok {
		return "optional"
	}
	return "unknown"
}

func hasValueKind(definitions []model.EnvVarDefinition, kind string) bool {
	for _, definition := range definitions {
		if definition.NormalizedValueKind == kind {
			return true
		}
	}
	return false
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	unique := []string{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return order.Strings(unique)
}

func isComposeFile(name string) bool {
	_, ok := composeFilenames[name]
	return ok
}

func isJavaScriptFile(filePath string) bool {
	_, ok := javaScriptExtensions[filepath.Ext(filePath)]
	return ok
}

func isGitHubActionsWorkflow(filePath string) bool {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) < 3 {
		return false
	}
	extension := filepath.Ext(filePath)
	return parts[len(parts)-3] == ".github" &&
		parts[len(parts)-2] == "workflows" &&
		(extension == ".yml" || extension == ".yaml")
}
