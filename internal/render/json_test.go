package render

import (
	"strings"
	"testing"

	"github.com/davisbuilds/envdiff/internal/model"
	"github.com/davisbuilds/envdiff/internal/testutil"
)

func TestJSONLeavesHTMLMetacharactersLiteral(t *testing.T) {
	envelope := model.NewJsonEnvelope(
		"compare",
		map[string]any{"left": "a.env", "right": "b.env"},
		map[string]any{"value": "postgres://h/db?a=1&b=2 <tag>"},
	)

	rendered, err := JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	if !strings.Contains(rendered, "?a=1&b=2 <tag>") {
		t.Fatalf("expected literal &, <, > in output; got:\n%s", rendered)
	}
	for _, escaped := range []string{"\\u0026", "\\u003c", "\\u003e"} {
		if strings.Contains(rendered, escaped) {
			t.Fatalf("output should not contain escaped %s; got:\n%s", escaped, rendered)
		}
	}
}

func TestJSONRendersCompareEnvelopeLikePythonGolden(t *testing.T) {
	envelope := model.NewJsonEnvelope(
		"compare",
		map[string]any{
			"left":  "tests/fixtures/compare/left.env",
			"right": "tests/fixtures/compare/right.env",
		},
		map[string]any{
			"differing_values": []map[string]any{
				{
					"left_kind":   "url",
					"left_value":  "postgres://localhost/db",
					"name":        "DATABASE_URL",
					"right_kind":  "integer",
					"right_value": "1",
				},
			},
			"duplicates_in_left":  []string{"DUP_KEY"},
			"duplicates_in_right": []string{},
			"left_path":           "tests/fixtures/compare/left.env",
			"missing_in_left":     []string{"FEATURE"},
			"missing_in_right":    []string{"DUP_KEY"},
			"right_path":          "tests/fixtures/compare/right.env",
			"warnings":            []string{},
		},
	)

	rendered, err := JSON(envelope)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	got := testutil.DecodeJSON(t, []byte(rendered))
	want := testutil.LoadGoldenJSON(t, "compare-basic.json")

	testutil.AssertJSONEqual(t, got, want)
}

func TestJSONRendersTopLevelKeysInStableOrder(t *testing.T) {
	rendered, err := JSON(model.NewJsonEnvelope("scan", map[string]any{"path": "."}, nil))
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	wantPrefix := "{\n  \"data\": {},\n  \"findings\": [],\n  \"inputs\": {"
	if len(rendered) < len(wantPrefix) || rendered[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("rendered prefix = %q, want %q", rendered[:len(wantPrefix)], wantPrefix)
	}
}
