package serverlog

import (
	"strings"
	"testing"
)

func TestTextFormat_Format(t *testing.T) {
	tf := NewTextFormat(4)

	result := tf.Format(InfoLevel, "test message", M{"key1": "val1", "key2": "val2"})

	// Should contain the level indicator
	if !strings.Contains(result, "INFO") {
		t.Errorf("expected format to contain 'INFO', got: %s", result)
	}
	// Should contain the message
	if !strings.Contains(result, "test message") {
		t.Errorf("expected format to contain 'test message', got: %s", result)
	}

	// Keys should be sorted, so key1 comes before key2
	key1Idx := strings.Index(result, "key1")
	key2Idx := strings.Index(result, "key2")
	if key1Idx == -1 {
		t.Error("expected format to contain 'key1'")
	}
	if key2Idx == -1 {
		t.Error("expected format to contain 'key2'")
	}
	if key1Idx > key2Idx {
		t.Errorf("keys should be sorted: key1=%d, key2=%d", key1Idx, key2Idx)
	}
}

func TestTextFormat_Format_EmptyData(t *testing.T) {
	tf := NewTextFormat(4)
	result := tf.Format(DebugLevel, "debug msg", M{})

	if !strings.Contains(result, "DEBU") {
		t.Errorf("expected format to contain 'DEBU', got: %s", result)
	}
	if !strings.Contains(result, "debug msg") {
		t.Errorf("expected format to contain 'debug msg', got: %s", result)
	}
}

func TestTextFormat_Format_NilData(t *testing.T) {
	tf := NewTextFormat(4)
	result := tf.Format(ErrorLevel, "error msg", nil)

	if !strings.Contains(result, "ERRO") {
		t.Errorf("expected format to contain 'ERRO', got: %s", result)
	}
	if !strings.Contains(result, "error msg") {
		t.Errorf("expected format to contain 'error msg', got: %s", result)
	}
}

func TestTextFormat_Format_DifferentLevels(t *testing.T) {
	tf := NewTextFormat(4)

	tests := []struct {
		level     Level
		prefix    string
	}{
		{PanicLevel, "PANI"},
		{FatalLevel, "FATA"},
		{ErrorLevel, "ERRO"},
		{WarnLevel, "WARN"},
		{InfoLevel, "INFO"},
		{DebugLevel, "DEBU"},
		{TraceLevel, "TRAC"},
	}

	for _, tt := range tests {
		result := tf.Format(tt.level, "msg", M{})
		if !strings.Contains(result, tt.prefix) {
			t.Errorf("expected %s to contain %q", tt.level, tt.prefix)
		}
	}
}

func TestTextFormat_Format_NonStringValues(t *testing.T) {
	tf := NewTextFormat(4)
	result := tf.Format(InfoLevel, "msg", M{"count": 42, "active": true})

	if !strings.Contains(result, "count=") {
		t.Error("expected 'count' key in output")
	}
	if !strings.Contains(result, "active=") {
		t.Error("expected 'active' key in output")
	}
}
