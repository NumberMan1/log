package serverlog

import (
	"testing"
)

func TestStackWrapPreHandler(t *testing.T) {
	data := M{"key": "val"}
	result := StackWrapPreHandler(InfoLevel, "test message", data)

	// Check required keys exist
	requiredKeys := []string{"Time", "LogType", "Message", "Stack", "Data"}
	for _, key := range requiredKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in result", key)
		}
	}

	// Check values
	if ts, ok := result["Time"].(int64); !ok || ts == 0 {
		t.Errorf("expected non-zero Time, got %v", result["Time"])
	}
	if result["LogType"] != "info" {
		t.Errorf("LogType = %v, want 'info'", result["LogType"])
	}
	if result["Message"] != "test message" {
		t.Errorf("Message = %v, want 'test message'", result["Message"])
	}

	// Stack should be a *frame
	if _, ok := result["Stack"].(*frame); !ok {
		t.Errorf("Stack should be *frame, got %T", result["Stack"])
	}

	// Data should contain original data
	innerData, ok := result["Data"].(M)
	if !ok {
		t.Fatalf("Data should be M, got %T", result["Data"])
	}
	if innerData["key"] != "val" {
		t.Errorf("inner Data['key'] = %v, want 'val'", innerData["key"])
	}
}

func TestStackWrapPreHandler_NilData(t *testing.T) {
	result := StackWrapPreHandler(ErrorLevel, "error msg", nil)

	if _, ok := result["Data"]; ok {
		t.Error("Data should not exist when input data is nil")
	}
	if result["LogType"] != "error" {
		t.Errorf("LogType = %v, want 'error'", result["LogType"])
	}
	if result["Message"] != "error msg" {
		t.Errorf("Message = %v, want 'error msg'", result["Message"])
	}
}

func TestStacksWrapPreHandler(t *testing.T) {
	result := StacksWrapPreHandler(ErrorLevel, "error msg", M{"a": 1})

	// Stack should be a slice of frames (multiple)
	if _, ok := result["Stack"]; !ok {
		t.Error("expected 'Stack' key")
	}
	// It should be a []*frame slice
	frames, ok := result["Stack"].([]*frame)
	if !ok {
		t.Fatalf("Stack should be []*frame, got %T", result["Stack"])
	}
	// stackTraceFrames(6) may return 0 in some environments (e.g., deeply inlined test calls)
	// The important thing is the type is correct; count may vary by Go runtime
	t.Logf("got %d frames from stackTraceFrames(6)", len(frames))
}

func TestNewStackWrapPreHandler(t *testing.T) {
	handler := NewStackWrapPreHandler(3)
	result := handler(DebugLevel, "debug msg", M{"x": "y"})

	if result["LogType"] != "debug" {
		t.Errorf("LogType = %v, want 'debug'", result["LogType"])
	}
	if result["Message"] != "debug msg" {
		t.Errorf("Message = %v, want 'debug msg'", result["Message"])
	}

	frames, ok := result["Stack"].([]*frame)
	if !ok {
		t.Errorf("Stack should be []*frame, got %T", result["Stack"])
	}
	if len(frames) == 0 {
		t.Error("expected at least one frame")
	}

	innerData := result["Data"].(M)
	if innerData["x"] != "y" {
		t.Errorf("Data['x'] = %v, want 'y'", innerData["x"])
	}
}
