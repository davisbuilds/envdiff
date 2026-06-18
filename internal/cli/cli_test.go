package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
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

	code := Run([]string{"compare", "left.env", "right.env"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "compare is not implemented") {
		t.Fatalf("stderr = %q, want unimplemented message", stderr.String())
	}
}

func TestRunScanJSONMatchesGolden(t *testing.T) {
	t.Chdir(testutil.RepoRoot(t))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"scan", "tests/fixtures/repos/simple_repo", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	got := testutil.DecodeJSON(t, stdout.Bytes())
	got = testutil.NormalizeJSONValue(got, testutil.DefaultPathReplacements(t))
	want := testutil.LoadGoldenJSON(t, "scan-simple-repo.json")
	testutil.AssertJSONEqual(t, got, want)
}
