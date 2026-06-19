package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/davisbuilds/envdiff/internal/analyzers"
	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/render"
)

var commandNames = []string{"compare", "generate", "matrix", "scan", "doctor"}

// Run executes the envdiff CLI and returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return 0
	}

	command := args[0]
	if !isKnownCommand(command) {
		fmt.Fprintf(stderr, "unknown command %q\n\n", command)
		printHelp(stderr)
		return 1
	}
	if command == "compare" {
		return runCompare(args[1:], stdout, stderr)
	}
	if command == "matrix" {
		return runMatrix(args[1:], stdout, stderr)
	}
	if command == "scan" {
		return runScan(args[1:], stdout, stderr)
	}
	if command == "generate" {
		return runGenerate(args[1:], stdout, stderr)
	}
	if command == "doctor" {
		return runDoctor(args[1:], stdout, stderr)
	}

	fmt.Fprintf(stderr, "%s is not implemented in the Go port yet\n", command)
	return 1
}

func printHelp(output io.Writer) {
	fmt.Fprintln(output, "Analyze repository environment contracts deterministically.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff <command> [options]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Commands:")
	for _, name := range commandNames {
		fmt.Fprintf(output, "  %s\n", name)
	}
}

func isKnownCommand(command string) bool {
	for _, name := range commandNames {
		if command == name {
			return true
		}
	}
	return false
}

func runScan(args []string, stdout io.Writer, stderr io.Writer) int {
	path := ""
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printScanHelp(stdout)
			return 0
		case "--json":
			jsonOutput = true
		default:
			if path != "" {
				fmt.Fprintf(stderr, "scan accepts exactly one path\n")
				return 1
			}
			path = arg
		}
	}
	if path == "" {
		fmt.Fprintf(stderr, "scan requires a repository path\n")
		return 1
	}

	result, err := analyzers.ScanRepository(path)
	if err != nil {
		fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 1
	}

	if jsonOutput {
		envelope := model.NewJsonEnvelope(
			"scan",
			map[string]any{"path": path},
			result,
		)
		rendered, err := render.JSON(envelope)
		if err != nil {
			fmt.Fprintf(stderr, "render JSON: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, rendered)
		return 0
	}

	fmt.Fprintln(stdout, render.ScanResult(result))
	return 0
}

func printScanHelp(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff scan <path> [--json]")
}

func runCompare(args []string, stdout io.Writer, stderr io.Writer) int {
	paths := []string{}
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printCompareHelp(stdout)
			return 0
		case "--json":
			jsonOutput = true
		default:
			paths = append(paths, arg)
		}
	}
	if len(paths) != 2 {
		fmt.Fprintf(stderr, "compare requires left and right dotenv files\n")
		return 1
	}

	result, err := analyzers.CompareDotenvFiles(paths[0], paths[1])
	if err != nil {
		fmt.Fprintf(stderr, "compare failed: %v\n", err)
		return 1
	}

	if jsonOutput {
		return printJSON(
			model.NewJsonEnvelope(
				"compare",
				map[string]any{"left": paths[0], "right": paths[1]},
				result,
			),
			stdout,
			stderr,
		)
	}

	fmt.Fprintln(stdout, render.CompareResult(result))
	return 0
}

func runMatrix(args []string, stdout io.Writer, stderr io.Writer) int {
	paths := []string{}
	showAll := false
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printMatrixHelp(stdout)
			return 0
		case "--show-all":
			showAll = true
		case "--json":
			jsonOutput = true
		default:
			paths = append(paths, arg)
		}
	}
	if len(paths) < 2 {
		fmt.Fprintf(stderr, "matrix requires at least two dotenv files\n")
		return 2
	}

	result, err := analyzers.MatrixDotenvFiles(paths, showAll)
	if err != nil {
		fmt.Fprintf(stderr, "matrix failed: %v\n", err)
		return 1
	}

	if jsonOutput {
		return printJSON(
			model.NewJsonEnvelope(
				"matrix",
				map[string]any{"paths": paths, "show_all": showAll},
				result,
			),
			stdout,
			stderr,
		)
	}

	fmt.Fprintln(stdout, render.MatrixResult(result))
	return 0
}

func printJSON(envelope model.JsonEnvelope, stdout io.Writer, stderr io.Writer) int {
	rendered, err := render.JSON(envelope)
	if err != nil {
		fmt.Fprintf(stderr, "render JSON: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, rendered)
	return 0
}

func printCompareHelp(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff compare <left> <right> [--json]")
}

func printMatrixHelp(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff matrix <paths...> [--show-all] [--json]")
}

func runGenerate(args []string, stdout io.Writer, stderr io.Writer) int {
	path := ""
	annotate := false
	check := false
	jsonOutput := false
	var outputPath *string

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--help", "-h":
			printGenerateHelp(stdout)
			return 0
		case "--annotate":
			annotate = true
		case "--check":
			check = true
		case "--json":
			jsonOutput = true
		case "--output":
			if index+1 >= len(args) {
				fmt.Fprintf(stderr, "--output requires a path\n")
				return 1
			}
			value := args[index+1]
			outputPath = &value
			index++
		default:
			if path != "" {
				fmt.Fprintf(stderr, "generate accepts exactly one path\n")
				return 1
			}
			path = arg
		}
	}
	if path == "" {
		fmt.Fprintf(stderr, "generate requires a repository path\n")
		return 1
	}

	scanResult, err := analyzers.ScanRepository(path)
	if err != nil {
		fmt.Fprintf(stderr, "generate scan failed: %v\n", err)
		return 1
	}
	result := analyzers.GenerateExampleFile(scanResult, annotate)

	var checkResult map[string]any
	if check {
		checkResult, err = analyzers.CheckGeneratedExample(
			scanResult.RootPath,
			result["generated_text"].(string),
			outputPath,
		)
		if err != nil {
			fmt.Fprintf(stderr, "generate check failed: %v\n", err)
			return 1
		}
	}

	var writtenPath *string
	if outputPath != nil && !check {
		written, err := analyzers.WriteGeneratedExample(*outputPath, result["generated_text"].(string))
		if err != nil {
			fmt.Fprintf(stderr, "generate write failed: %v\n", err)
			return 1
		}
		writtenPath = &written
	}

	// A failed --check drives the exit code regardless of output format, so a
	// `generate --check --json` gate is just as enforceable as the human path.
	driftExit := 0
	if checkResult != nil && !checkResult["matches"].(bool) {
		driftExit = 2
	}

	if jsonOutput {
		result["output_path"] = nil
		if writtenPath != nil {
			result["output_path"] = *writtenPath
		}
		result["check"] = checkResult
		if code := printJSON(
			model.NewJsonEnvelope(
				"generate",
				map[string]any{
					"annotate": annotate,
					"check":    check,
					"output":   outputPathValue(outputPath),
					"path":     path,
				},
				result,
			),
			stdout,
			stderr,
		); code != 0 {
			return code
		}
		return driftExit
	}

	if checkResult != nil {
		checkPath := checkResult["target_path"].(string)
		checkMatches := checkResult["matches"].(bool)
		fmt.Fprintln(
			stdout,
			render.GenerateResult(
				result["variable_count"].(int),
				nil,
				annotate,
				&checkPath,
				&checkMatches,
			),
		)
		return driftExit
	}
	if writtenPath != nil {
		fmt.Fprintln(
			stdout,
			render.GenerateResult(result["variable_count"].(int), writtenPath, annotate, nil, nil),
		)
		return 0
	}

	fmt.Fprint(stdout, result["generated_text"].(string))
	return 0
}

func outputPathValue(outputPath *string) any {
	if outputPath == nil {
		return nil
	}
	return *outputPath
}

func printGenerateHelp(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff generate <path> [--annotate] [--check] [--output <path>] [--json]")
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	path := ""
	failOn := "error"
	var baselinePath *string
	var writeBaselinePath *string
	var ignoreFilePath *string
	jsonOutput := false

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--help", "-h":
			printDoctorHelp(stdout)
			return 0
		case "--json":
			jsonOutput = true
		case "--fail-on":
			value, ok := nextOptionValue(args, &index)
			if !ok {
				fmt.Fprintf(stderr, "--fail-on requires a severity\n")
				return 1
			}
			failOn = value
		case "--baseline":
			value, ok := nextOptionValue(args, &index)
			if !ok {
				fmt.Fprintf(stderr, "--baseline requires a path\n")
				return 1
			}
			baselinePath = &value
		case "--write-baseline":
			value, ok := nextOptionValue(args, &index)
			if !ok {
				fmt.Fprintf(stderr, "--write-baseline requires a path\n")
				return 1
			}
			writeBaselinePath = &value
		case "--ignore-file":
			value, ok := nextOptionValue(args, &index)
			if !ok {
				fmt.Fprintf(stderr, "--ignore-file requires a path\n")
				return 1
			}
			ignoreFilePath = &value
		default:
			if path != "" {
				fmt.Fprintf(stderr, "doctor accepts exactly one path\n")
				return 1
			}
			path = arg
		}
	}
	if path == "" {
		fmt.Fprintf(stderr, "doctor requires a repository path\n")
		return 1
	}

	if _, err := analyzers.ShouldFail(model.SummaryCounts{}, failOn); err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		return 2
	}

	scanResult, err := analyzers.ScanRepository(path)
	if err != nil {
		fmt.Fprintf(stderr, "doctor scan failed: %v\n", err)
		return 1
	}
	findings := analyzers.DoctorRepository(scanResult)

	suppressionKeys := map[string]struct{}{}
	baselineEntryCount := 0
	if baselinePath != nil {
		snapshot, err := analyzers.LoadBaselineSnapshot(*baselinePath)
		if err != nil {
			fmt.Fprintf(stderr, "load baseline failed: %v\n", err)
			return 1
		}
		baselineEntryCount = len(snapshot.Entries)
		for _, entry := range snapshot.Entries {
			suppressionKeys[entry.SuppressionKey] = struct{}{}
		}
	}

	defaultIgnorePath := filepath.Join(path, ".envdiffignore")
	if ignoreFilePath != nil {
		keys, err := analyzers.LoadIgnoreKeys(*ignoreFilePath)
		if err != nil {
			fmt.Fprintf(stderr, "load ignore file failed: %v\n", err)
			return 1
		}
		for key := range keys {
			suppressionKeys[key] = struct{}{}
		}
	} else if info, err := os.Stat(defaultIgnorePath); err == nil && !info.IsDir() {
		ignoreFilePath = &defaultIgnorePath
		keys, err := analyzers.LoadIgnoreKeys(defaultIgnorePath)
		if err != nil {
			fmt.Fprintf(stderr, "load ignore file failed: %v\n", err)
			return 1
		}
		for key := range keys {
			suppressionKeys[key] = struct{}{}
		}
	}

	activeFindings, suppressedFindings := analyzers.ApplySuppressions(findings, suppressionKeys)
	summary := analyzers.SummarizeFindings(activeFindings)
	var baselineWritten *string
	if writeBaselinePath != nil {
		if _, err := analyzers.WriteBaselineSnapshot(*writeBaselinePath, findings); err != nil {
			fmt.Fprintf(stderr, "write baseline failed: %v\n", err)
			return 1
		}
		baselineWritten = writeBaselinePath
	}

	if jsonOutput {
		envelope := model.NewJsonEnvelope(
			"doctor",
			map[string]any{
				"baseline":       outputPathValue(baselinePath),
				"fail_on":        failOn,
				"ignore_file":    outputPathValue(ignoreFilePath),
				"path":           path,
				"write_baseline": outputPathValue(writeBaselinePath),
			},
			map[string]any{
				"filtering": map[string]any{
					"baseline_entries": baselineEntryCount,
					"baseline_written": outputPathValue(baselineWritten),
					"suppressed_count": len(suppressedFindings),
				},
				"scan":                scanResult,
				"suppressed_findings": suppressedFindings,
			},
		)
		envelope.Findings = activeFindings
		envelope.Summary = summary
		if code := printJSON(envelope, stdout, stderr); code != 0 {
			return code
		}
	} else {
		fmt.Fprintln(
			stdout,
			render.DoctorResult(
				scanResult.RootPath,
				activeFindings,
				len(suppressedFindings),
				baselineWritten,
			),
		)
	}

	if writeBaselinePath != nil {
		return 0
	}
	shouldFail, err := analyzers.ShouldFail(summary, failOn)
	if err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		return 1
	}
	if shouldFail {
		return 2
	}
	return 0
}

func nextOptionValue(args []string, index *int) (string, bool) {
	if *index+1 >= len(args) {
		return "", false
	}
	*index = *index + 1
	return args[*index], true
}

func printDoctorHelp(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  envdiff doctor <path> [--fail-on <severity>] [--baseline <path>] [--write-baseline <path>] [--ignore-file <path>] [--json]")
}
