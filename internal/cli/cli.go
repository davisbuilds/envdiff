package cli

import (
	"fmt"
	"io"

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
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Use the Python launcher ./envdiff until Go parity is complete.")
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

	fmt.Fprintf(stdout, "Scan root: %s\n", result.RootPath)
	fmt.Fprintf(stdout, "Definitions: %d\n", len(result.Definitions))
	fmt.Fprintf(stdout, "Usages: %d\n", len(result.Usages))
	fmt.Fprintf(stdout, "Contracts: %d\n", len(result.Contracts))
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
		return 1
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

	if jsonOutput {
		result["output_path"] = nil
		if writtenPath != nil {
			result["output_path"] = *writtenPath
		}
		result["check"] = checkResult
		return printJSON(
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
		)
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
		if !checkMatches {
			return 2
		}
		return 0
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
