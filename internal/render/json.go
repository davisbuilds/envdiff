package render

import (
	"bytes"
	"encoding/json"

	"github.com/davisbuilds/envdiff/internal/model"
)

// MarshalIndent encodes v as two-space-indented JSON without a trailing newline.
// Unlike json.MarshalIndent it leaves <, >, and & literal rather than escaping
// them to \u00XX — this output is data for humans and tooling, not HTML.
func MarshalIndent(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func JSON(envelope model.JsonEnvelope) (string, error) {
	payload, err := MarshalIndent(envelope)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
