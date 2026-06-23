package serverlog

import (
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
		{ErrorLevel, "error"},
		{WarnLevel, "warning"},
		{InfoLevel, "info"},
		{DebugLevel, "debug"},
		{TraceLevel, "trace"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestLevel_String_Unknown(t *testing.T) {
	l := Level(255)
	if got := l.String(); got != "unknown" {
		t.Errorf("unknown Level.String() = %q, want %q", got, "unknown")
	}
}

func TestLevel_MarshalText(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
		{ErrorLevel, "error"},
		{WarnLevel, "warning"},
		{InfoLevel, "info"},
		{DebugLevel, "debug"},
		{TraceLevel, "trace"},
	}

	for _, tt := range tests {
		b, err := tt.level.MarshalText()
		if err != nil {
			t.Errorf("Level(%d).MarshalText() error = %v", tt.level, err)
			continue
		}
		if string(b) != tt.expected {
			t.Errorf("Level(%d).MarshalText() = %q, want %q", tt.level, string(b), tt.expected)
		}
	}
}

func TestLevel_MarshalText_Invalid(t *testing.T) {
	l := Level(255)
	_, err := l.MarshalText()
	if err == nil {
		t.Error("expected error for invalid level, got nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		wantErr  bool
	}{
		{"panic", PanicLevel, false},
		{"PANIC", PanicLevel, false},
		{"fatal", FatalLevel, false},
		{"error", ErrorLevel, false},
		{"warn", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"WARNING", WarnLevel, false},
		{"info", InfoLevel, false},
		{"debug", DebugLevel, false},
		{"trace", TraceLevel, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseLevel(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseLevel(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestAllLevels(t *testing.T) {
	expected := []Level{PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel}
	if len(AllLevels) != len(expected) {
		t.Errorf("AllLevels length = %d, want %d", len(AllLevels), len(expected))
	}
	for i, l := range expected {
		if AllLevels[i] != l {
			t.Errorf("AllLevels[%d] = %d, want %d", i, AllLevels[i], l)
		}
	}
}
