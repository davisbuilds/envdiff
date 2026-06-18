package model

import (
	"encoding/json"
	"testing"
)

func TestJsonEnvelopeDefaultsMatchSchemaContract(t *testing.T) {
	payload, err := json.Marshal(NewJsonEnvelope("scan", map[string]any{"path": "."}, nil))
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}

	meta := decoded["meta"].(map[string]any)
	if meta["schema_version"] != "1" {
		t.Fatalf("schema_version = %v, want 1", meta["schema_version"])
	}
	if len(decoded["data"].(map[string]any)) != 0 {
		t.Fatalf("data = %#v, want empty object", decoded["data"])
	}
	if len(decoded["findings"].([]any)) != 0 {
		t.Fatalf("findings = %#v, want empty array", decoded["findings"])
	}
}

func TestNestedSlicesMarshalAsEmptyArrays(t *testing.T) {
	contract := EnvVarContract{Name: "DATABASE_URL"}
	payload, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("marshal contract: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode contract: %v", err)
	}

	for _, key := range []string{"aliases", "definitions", "resolution_notes", "status", "usages"} {
		values, ok := decoded[key].([]any)
		if !ok {
			t.Fatalf("%s = %#v, want empty array", key, decoded[key])
		}
		if len(values) != 0 {
			t.Fatalf("%s = %#v, want empty array", key, values)
		}
	}
	if decoded["requiredness"] != "unknown" {
		t.Fatalf("requiredness = %v, want unknown", decoded["requiredness"])
	}
}

func TestBaselineSnapshotDefaultsSchemaVersion(t *testing.T) {
	payload, err := json.Marshal(BaselineSnapshot{})
	if err != nil {
		t.Fatalf("marshal baseline snapshot: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode baseline snapshot: %v", err)
	}

	if decoded["schema_version"] != "1" {
		t.Fatalf("schema_version = %v, want 1", decoded["schema_version"])
	}
	if len(decoded["entries"].([]any)) != 0 {
		t.Fatalf("entries = %#v, want empty array", decoded["entries"])
	}
}
