package serverlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
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

type jsonTestMap map[string]any
type jsonTestErrorSlice []error
type jsonTestBytes []byte

type jsonTestNilError struct{ message string }

// 解引用接收者，确保 typed nil 不会被误调用。
func (err *jsonTestNilError) Error() string { return err.message }

type jsonTestMarshaler map[string]any

func (jsonTestMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"custom":true}`), nil
}

type jsonTestErrorMarshaler struct{}

func (*jsonTestErrorMarshaler) Error() string                { return "error takes precedence" }
func (*jsonTestErrorMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"custom"`), nil }

func assertJSONValue(t *testing.T, got string, want any) {
	t.Helper()
	expected, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(expected) {
		t.Fatalf("JSON = %s, want %s", got, expected)
	}
}

func TestJsonFormatPreservesOrdinaryJSON(t *testing.T) {
	for _, input := range jsonBenchmarkInputs() {
		t.Run(input.name, func(t *testing.T) {
			assertJSONValue(t, NewJsonFormat().Format(InfoLevel, "", input.data), input.data)
		})
	}
	data := M{
		"nil_map": M(nil), "nil_slice": []any(nil), "empty_map": M{}, "empty_slice": []any{},
		"bytes": []byte{1, 2, 3}, "nil_bytes": []byte(nil), "number": json.Number("1234567890123456789"),
		"time": time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "pointer": new(int),
		"integers": []int{1, 2, 3}, "array": [2]string{"first", "second"},
	}
	assertJSONValue(t, NewJsonFormat().Format(InfoLevel, "", data), data)
}

func TestJsonFormatConvertsErrorsWithoutMutatingInput(t *testing.T) {
	original := errors.New("unavailable")
	wrapped := fmt.Errorf("load: %w", original)
	shared := M{"untouched": true}
	items := []any{shared, original, M{"error": wrapped}, "tail"}
	typedMap := jsonTestMap{"error": original}
	typedSlice := jsonTestErrorSlice{original, nil}
	array := [2]error{original, nil}
	data := M{"items": items, "typed_map": typedMap, "typed_slice": typedSlice, "array": array}
	assertJSONValue(t, NewJsonFormat().Format(ErrorLevel, "", data), M{
		"items":     []any{M{"untouched": true}, "unavailable", M{"error": "load: unavailable"}, "tail"},
		"typed_map": M{"error": "unavailable"}, "typed_slice": []any{"unavailable", nil}, "array": []any{"unavailable", nil},
	})
	if items[1] != original || items[2].(M)["error"] != wrapped || typedMap["error"] != original || typedSlice[0] != original || array[0] != original {
		t.Fatal("formatting changed an input error")
	}
	if !reflect.DeepEqual(shared, M{"untouched": true}) {
		t.Fatal("formatting changed an unaffected branch")
	}
}

func TestJsonFormatCustomTypesAndTypedNil(t *testing.T) {
	var nilError *jsonTestNilError
	data := M{
		"error": nilError, "dual_nil": (*jsonTestErrorMarshaler)(nil),
		"nil_map": jsonTestMap(nil), "nil_slice": jsonTestErrorSlice(nil),
		"nil_func": (func())(nil), "nil_channel": (chan int)(nil),
		"custom":     jsonTestMarshaler{"error": errors.New("hidden")},
		"nil_custom": jsonTestMarshaler(nil), "dual": &jsonTestErrorMarshaler{},
		// 命名 byte 切片维持既有的数组输出，普通 []byte 则沿用标准库的 base64。
		"named_bytes": jsonTestBytes{1, 2}, "bytes": []byte{1, 2},
	}
	assertJSONValue(t, NewJsonFormat().Format(InfoLevel, "", data), M{
		"error": nil, "dual_nil": nil, "nil_map": nil, "nil_slice": nil,
		"nil_func": nil, "nil_channel": nil, "nil_custom": nil,
		"custom": M{"custom": true}, "dual": "error takes precedence",
		"named_bytes": []int{1, 2}, "bytes": []byte{1, 2},
	})
}

func TestJsonFormatDepthLimitAndCycles(t *testing.T) {
	format := NewJsonFormat()
	wrap := func(value any, depth int) M {
		for range depth {
			value = M{"next": value}
		}
		return value.(M)
	}
	for _, depth := range []int{31, 32, 33} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			want := any("leaf")
			if depth >= 32 {
				want = "[max depth]"
			}
			assertJSONValue(t, format.Format(InfoLevel, "", wrap("leaf", depth)), wrap(want, min(depth, 32)))
		})
	}
	for _, nilValue := range []any{nil, M(nil), []any(nil), []byte(nil), (*jsonTestNilError)(nil)} {
		assertJSONValue(t, format.Format(InfoLevel, "", wrap(nilValue, 32)), wrap(nil, 32))
	}
	for _, slice := range []any{[]string{"first", "second"}, []int{1, 2}, []json.Number{"1", "2"}} {
		assertJSONValue(t, format.Format(InfoLevel, "", wrap(slice, 30)), wrap(slice, 30))
		assertJSONValue(t, format.Format(InfoLevel, "", wrap(slice, 31)), wrap([]any{"[max depth]", "[max depth]"}, 31))
		assertJSONValue(t, format.Format(InfoLevel, "", wrap(slice, 32)), wrap("[max depth]", 32))
	}
	assertJSONValue(t, format.Format(InfoLevel, "", wrap([]string(nil), 31)), wrap(nil, 31))
	cyclicMap := M{}
	cyclicMap["next"] = cyclicMap
	assertJSONValue(t, format.Format(InfoLevel, "", cyclicMap), wrap("[max depth]", 32))
	if _, ok := cyclicMap["next"].(M); !ok {
		t.Fatal("formatting modified the cyclic map")
	}
	cyclicSlice := make([]any, 1)
	cyclicSlice[0] = cyclicSlice
	var want any = "[max depth]"
	for range 31 {
		want = []any{want}
	}
	assertJSONValue(t, format.Format(InfoLevel, "", M{"cycle": cyclicSlice}), M{"cycle": want})
	if _, ok := cyclicSlice[0].([]any); !ok {
		t.Fatal("formatting modified the cyclic slice")
	}
}

func TestJsonFormatConcurrentSharedPayload(t *testing.T) {
	format := NewJsonFormat()
	err := errors.New("unavailable")
	data := M{"items": []any{M{"error": err}, M{"value": true}}}
	want := `{"items":[{"error":"unavailable"},{"value":true}]}`
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 100 {
				if got := format.Format(ErrorLevel, "", data); got != want {
					t.Errorf("JSON = %s, want %s", got, want)
					return
				}
			}
		})
	}
	wg.Wait()
	if data["items"].([]any)[0].(M)["error"] != err {
		t.Fatal("concurrent formatting changed the input error")
	}
}
