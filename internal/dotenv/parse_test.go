package dotenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestParsePreservesDuplicatesAndValueKinds(t *testing.T) {
	result, err := Parse(testutil.FixturePath(t, "dotenv", "basic.env"))
	if err != nil {
		t.Fatalf("parse dotenv: %v", err)
	}

	names := make([]string, 0, len(result.Definitions))
	kinds := map[string]string{}
	for _, definition := range result.Definitions {
		names = append(names, definition.Name)
		kinds[definition.Name] = definition.NormalizedValueKind
	}

	wantNames := []string{"DATABASE_URL", "DEBUG", "EMPTY", "QUOTED"}
	for index, want := range wantNames {
		if names[index] != want {
			t.Fatalf("names = %v, want %v", names, wantNames)
		}
	}
	if kinds["DATABASE_URL"] != "url" {
		t.Fatalf("DATABASE_URL kind = %s, want url", kinds["DATABASE_URL"])
	}
	if kinds["DEBUG"] != "boolean" {
		t.Fatalf("DEBUG kind = %s, want boolean", kinds["DEBUG"])
	}
	if kinds["EMPTY"] != "placeholder" {
		t.Fatalf("EMPTY kind = %s, want placeholder", kinds["EMPTY"])
	}
	if kinds["QUOTED"] != "string" {
		t.Fatalf("QUOTED kind = %s, want string", kinds["QUOTED"])
	}
}

func TestParseMarksDuplicateDefinitions(t *testing.T) {
	result, err := Parse(testutil.FixturePath(t, "dotenv", "duplicates.env"))
	if err != nil {
		t.Fatalf("parse dotenv: %v", err)
	}

	if len(result.Definitions) != 2 {
		t.Fatalf("definition count = %d, want 2", len(result.Definitions))
	}
	if result.Definitions[0].IsDuplicate {
		t.Fatal("first definition should not be duplicate")
	}
	if !result.Definitions[1].IsDuplicate {
		t.Fatal("second definition should be duplicate")
	}
}

func TestParseRecordsWarningsForUnsupportedSyntax(t *testing.T) {
	path := testutil.FixturePath(t, "dotenv", "malformed.env")
	result, err := Parse(path)
	if err != nil {
		t.Fatalf("parse dotenv: %v", err)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(result.Warnings))
	}
	if !strings.Contains(result.Warnings[0], filepath.ToSlash(path)+":2: unsupported dotenv syntax") {
		t.Fatalf("warning = %q, want unsupported syntax with path and line", result.Warnings[0])
	}
}

func TestParseStripsMatchingSingleAndDoubleQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "SINGLE='hello world'\nDOUBLE=\"hello\"\nMISMATCH=\"hello'\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := Parse(path)
	if err != nil {
		t.Fatalf("parse dotenv: %v", err)
	}

	values := map[string]string{}
	for _, definition := range result.Definitions {
		values[definition.Name] = definition.Value
	}
	if values["SINGLE"] != "hello world" {
		t.Fatalf("SINGLE = %q, want hello world", values["SINGLE"])
	}
	if values["DOUBLE"] != "hello" {
		t.Fatalf("DOUBLE = %q, want hello", values["DOUBLE"])
	}
	if values["MISMATCH"] != "\"hello'" {
		t.Fatalf("MISMATCH = %q, want unmatched quotes preserved", values["MISMATCH"])
	}
}
