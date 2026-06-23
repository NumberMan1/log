package serverlog

import (
	"encoding/json"
	"testing"
)

func TestJsonFormat_Format(t *testing.T) {
	jf := NewJsonFormat()

	data := M{"key1": "val1", "key2": 42}
	result := jf.Format(InfoLevel, "msg", data)

	// Validate it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("expected valid JSON, got error: %v, output: %s", err, result)
	}
	if parsed["key1"] != "val1" {
		t.Errorf("key1 = %v, want 'val1'", parsed["key1"])
	}
	if parsed["key2"] != float64(42) { // JSON numbers unmarshal as float64
		t.Errorf("key2 = %v, want 42", parsed["key2"])
	}
}

func TestJsonFormat_Format_NilData(t *testing.T) {
	jf := NewJsonFormat()
	result := jf.Format(InfoLevel, "msg", nil)

	if result != "" {
		t.Errorf("expected empty string for nil data, got: %s", result)
	}
}

func TestJsonFormat_Format_EmptyData(t *testing.T) {
	jf := NewJsonFormat()
	result := jf.Format(InfoLevel, "msg", M{})

	if result != "{}" {
		t.Errorf("expected '{}' for empty data, got: %s", result)
	}
}

func TestJsonFormat_Format_ComplexData(t *testing.T) {
	jf := NewJsonFormat()

	data := M{
		"name": "test",
		"nested": M{
			"a": 1,
			"b": "hello",
		},
		"slice": []int{1, 2, 3},
	}
	result := jf.Format(InfoLevel, "msg", data)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v, output: %s", err, result)
	}
	if parsed["name"] != "test" {
		t.Errorf("name = %v, want 'test'", parsed["name"])
	}
	if parsed["slice"] == nil {
		t.Error("slice should not be nil")
	}
}

func TestJsonFormat_Format_UnmarshalableType(t *testing.T) {
	jf := NewJsonFormat()

	// Channel is not JSON-marshalable
	data := M{"ch": make(chan int)}
	result := jf.Format(InfoLevel, "msg", data)

	// Should produce an error JSON, not a panic
	if result == "" {
		t.Error("expected non-empty result for unmarshalable data")
	}
}
