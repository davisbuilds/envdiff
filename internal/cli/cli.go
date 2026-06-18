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
	if command == "scan" {
		return runScan(args[1:], stdout, stderr)
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
