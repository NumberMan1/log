package serverlog

import (
	"testing"
)

func TestOnlyMessageFormat_Format(t *testing.T) {
	omf := NewOnlyMessageFormat()

	tests := []struct {
		msg  string
		data M
	}{
		{"hello", M{"key": "val"}},
		{"", M{}},
		{"world", nil},
	}

	for _, tt := range tests {
		result := omf.Format(InfoLevel, tt.msg, tt.data)
		if result != tt.msg {
			t.Errorf("OnlyMessageFormat.Format() = %q, want %q", result, tt.msg)
		}
	}
}

func TestOnlyMessageFormat_Format_IgnoresLevel(t *testing.T) {
	omf := NewOnlyMessageFormat()

	// The format should return the message regardless of level
	for _, lv := range AllLevels {
		result := omf.Format(lv, "test", M{"a": 1})
		if result != "test" {
			t.Errorf("OnlyMessageFormat.Format(%v) = %q, want 'test'", lv, result)
		}
	}
}
