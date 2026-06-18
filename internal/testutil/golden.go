package testutil

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	RepoRootToken = "<REPO_ROOT>"
	TmpDirToken   = "<TMPDIR>"
)

// LoadGoldenJSON loads a committed JSON golden under tests/golden/json.
func LoadGoldenJSON(t testing.TB, name string) any {
	t.Helper()

	path := filepath.Join(RepoRoot(t), "tests", "golden", "json", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}

	return DecodeJSON(t, payload)
}

// DecodeJSON decodes JSON into values suitable for structural comparisons.
func DecodeJSON(t testing.TB, payload []byte) any {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	return value
}

// NormalizeJSONValue recursively normalizes local paths inside decoded JSON values.
func NormalizeJSONValue(value any, replacements map[string]string) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(typed))
		for key, child := range typed {
			normalized[key] = NormalizeJSONValue(child, replacements)
		}
		return normalized
	case []any:
		normalized := make([]any, len(typed))
		for index, child := range typed {
			normalized[index] = NormalizeJSONValue(child, replacements)
		}
		return normalized
	case string:
		return normalizeString(typed, replacements)
	default:
		return value
	}
}

// DefaultPathReplacements returns replacements for the current checkout.
func DefaultPathReplacements(t testing.TB) map[string]string {
	t.Helper()

	return map[string]string{
		filepath.ToSlash(RepoRoot(t)): RepoRootToken,
	}
}

// AssertJSONEqual compares decoded JSON values and reports a stable diff-friendly payload.
func AssertJSONEqual(t testing.TB, got any, want any) {
	t.Helper()

	if reflect.DeepEqual(got, want) {
		return
	}

	t.Fatalf("JSON mismatch\nwant:\n%s\n\ngot:\n%s", PrettyJSON(t, want), PrettyJSON(t, got))
}

// PrettyJSON renders decoded JSON values with deterministic map-key order.
func PrettyJSON(t testing.TB, value any) string {
	t.Helper()

	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	return string(payload)
}

func normalizeString(value string, replacements map[string]string) string {
	normalized := strings.ReplaceAll(value, "\\", "/")
	keys := make([]string, 0, len(replacements))
	for source := range replacements {
		keys = append(keys, source)
	}
	sort.Slice(keys, func(left int, right int) bool {
		return len(keys[left]) > len(keys[right])
	})

	for _, source := range keys {
		normalized = strings.ReplaceAll(normalized, filepath.ToSlash(source), replacements[source])
	}
	return normalized
}
