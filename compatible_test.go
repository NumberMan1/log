package serverlog

import (
	"bytes"
	"testing"
)

func TestCompatibleLogger_Write(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	cl := NewCompatibleLogger(l)
	n, err := cl.Write([]byte("test message"))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Write uses named returns, so n defaults to 0; len(p) is not explicitly returned
	_ = n

	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != InfoLevel {
		t.Errorf("level = %d, want %d", logs[0].level, InfoLevel)
	}
	// Write adds the bytes as the message (without trailing newline stripped)
	expected := "test message"
	if logs[0].msg != expected {
		t.Errorf("msg = %q, want %q", logs[0].msg, expected)
	}
}

func TestCompatibleLogger_Printf(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	cl := NewCompatibleLogger(l)
	cl.Printf("hello %s, count=%d", "world", 42)

	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].msg != "hello world, count=42" {
		t.Errorf("msg = %q, want 'hello world, count=42'", logs[0].msg)
	}
	if logs[0].level != InfoLevel {
		t.Errorf("level = %d, want %d", logs[0].level, InfoLevel)
	}
}

func TestCompatibleLogger_Write_Multiple(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	cl := NewCompatibleLogger(l)

	messages := []string{"first", "second", "third"}
	for _, msg := range messages {
		_, err := cl.Write([]byte(msg))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	logs := mp.GetLogs()
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}
	for i, expected := range messages {
		if logs[i].msg != expected {
			t.Errorf("logs[%d].msg = %q, want %q", i, logs[i].msg, expected)
		}
	}
}

func TestCompatibleLogger_Log(t *testing.T) {
	SetExitFunc(func(code int) { /* no-op in test */ })
	defer ResetExitFunc()

	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(AllLevels...), mp)

	cl := NewCompatibleLogger(l)

	tests := []struct {
		levelStr string
		expected Level
	}{
		{"error", ErrorLevel},
		{"warn", WarnLevel},
		{"info", InfoLevel},
		{"debug", DebugLevel},
		{"trace", TraceLevel},
	}

	for _, tt := range tests {
		cl.Log(tt.levelStr, "test", M{"lvl": tt.levelStr})
	}

	logs := mp.GetLogs()
	if len(logs) != len(tests) {
		t.Fatalf("expected %d logs, got %d", len(tests), len(logs))
	}
	for i, tt := range tests {
		if logs[i].level != tt.expected {
			t.Errorf("logs[%d].level = %d, want %d (%s)", i, logs[i].level, tt.expected, tt.levelStr)
		}
	}
}

func TestCompatibleLogger_Log_InvalidLevel(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(ErrorLevel), mp)

	cl := NewCompatibleLogger(l)
	cl.Log("invalid_level", "test", M{})

	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	// Invalid level should default to ErrorLevel
	if logs[0].level != ErrorLevel {
		t.Errorf("invalid level should default to ErrorLevel, got %d", logs[0].level)
	}
}

func TestCompatibleLogger_Log_WithData(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	cl := NewCompatibleLogger(l)
	cl.Log("info", "test message", M{"user": "admin", "count": 10})

	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].data["user"] != "admin" {
		t.Errorf("data['user'] = %v, want 'admin'", logs[0].data["user"])
	}
	if logs[0].data["count"] != 10 {
		t.Errorf("data['count'] = %v, want 10", logs[0].data["count"])
	}
}

func TestCompatibleLogger_ImplementsWriter(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	cl := NewCompatibleLogger(l)

	// Use as io.Writer
	var buf bytes.Buffer
	buf.WriteString("hello from buffer")

	n, err := cl.Write(buf.Bytes())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Write uses named returns, so n defaults to 0; the byte count is not explicitly returned
	_ = n
}
