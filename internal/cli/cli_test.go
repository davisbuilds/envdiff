package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsHelpForNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	for _, command := range commandNames {
		if !strings.Contains(output, command) {
			t.Fatalf("help output missing command %q:\n%s", command, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsUnimplementedCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"scan", "."}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "scan is not implemented") {
		t.Fatalf("stderr = %q, want unimplemented message", stderr.String())
	}
}
