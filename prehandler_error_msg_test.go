package serverlog

import (
	"fmt"
	"testing"
)

func TestErrorMsgPreHandler(t *testing.T) {
	err := fmt.Errorf("something went wrong")
	data := M{"error": err}
	result := ErrorMsgPreHandler(ErrorLevel, "msg", data)

	if result["error"] != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %v", result["error"])
	}
}

func TestErrorMsgPreHandler_ErrKey(t *testing.T) {
	err := fmt.Errorf("another error")
	data := M{"err": err}
	result := ErrorMsgPreHandler(ErrorLevel, "msg", data)

	if result["err"] != "another error" {
		t.Errorf("expected 'another error', got %v", result["err"])
	}
}

func TestErrorMsgPreHandler_NoErrorKey(t *testing.T) {
	data := M{"key": "val"}
	result := ErrorMsgPreHandler(ErrorLevel, "msg", data)

	if result["key"] != "val" {
		t.Errorf("data should be unchanged, got %v", result)
	}
}

func TestErrorMsgPreHandler_NilData(t *testing.T) {
	result := ErrorMsgPreHandler(ErrorLevel, "msg", nil)
	if result != nil {
		t.Errorf("expected nil for nil data, got %v", result)
	}
}

func TestErrorMsgPreHandler_NilError(t *testing.T) {
	data := M{"error": nil}
	result := ErrorMsgPreHandler(ErrorLevel, "msg", data)

	if result["error"] != "" {
		t.Errorf("expected empty string for nil error, got %v", result["error"])
	}
}

func TestErrorMsgPreHandler_NonErrorInterface(t *testing.T) {
	data := M{"error": 42}
	result := ErrorMsgPreHandler(ErrorLevel, "msg", data)

	if result["error"] != "42" {
		t.Errorf("expected '42', got %v", result["error"])
	}
}

func TestExtraErrorMsg(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, ""},
		{fmt.Errorf("test error"), "test error"},
		{42, "42"},
		{"hello", "hello"},
		{true, "true"},
	}

	for _, tt := range tests {
		got := extraErrorMsg(tt.input)
		if got != tt.expected {
			t.Errorf("extraErrorMsg(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
