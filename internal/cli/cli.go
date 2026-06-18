package cli

import (
	"fmt"
	"io"
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
