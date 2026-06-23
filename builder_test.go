package serverlog

import (
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.l == nil {
		t.Fatal("builder's logger is nil")
	}
}

func TestBuilder_AddCustomPrinter(t *testing.T) {
	b := NewBuilder()
	mp := &mockPrinter{}
	result := b.AddCustomPrinter(ErrorLevel, mp)

	// Should return self for chaining
	if result != b {
		t.Error("AddCustomPrinter should return self")
	}

	// The printer should be registered
	logger := b.Logger()
	if sw := logger.GetLevelSwitcherByPrinter(mp); sw == nil {
		t.Error("printer should be registered in logger")
	}

	// Verify level filtering: only ErrorLevel and below should be enabled
	ls := logger.GetLevelSwitcherByPrinter(mp)
	if !ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
	if !ls.IsEnabled(FatalLevel) {
		t.Error("FatalLevel should be enabled")
	}
	if !ls.IsEnabled(PanicLevel) {
		t.Error("PanicLevel should be enabled")
	}
	if ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should not be enabled")
	}
}

func TestBuilder_Logger(t *testing.T) {
	b := NewBuilder()
	l := b.Logger()
	if l == nil {
		t.Fatal("Logger() returned nil")
	}
	if l.maxLevel != TraceLevel {
		t.Errorf("maxLevel = %d, want %d", l.maxLevel, TraceLevel)
	}
}

func TestBuilder_ChainedCalls(t *testing.T) {
	b := NewBuilder()
	mp1 := &mockPrinter{}
	mp2 := &mockPrinter{}

	result := b.AddCustomPrinter(InfoLevel, mp1).
		AddCustomPrinter(ErrorLevel, mp2)

	if result != b {
		t.Error("chained calls should return self")
	}

	logger := b.Logger()
	if logger.GetLevelSwitcherByPrinter(mp1) == nil {
		t.Error("mp1 should be registered")
	}
	if logger.GetLevelSwitcherByPrinter(mp2) == nil {
		t.Error("mp2 should be registered")
	}
}

func TestBuilder_LoggerLogging(t *testing.T) {
	b := NewBuilder()
	mp := &mockPrinter{}
	b.AddCustomPrinter(InfoLevel, mp)

	logger := b.Logger()
	logger.Info("builder test", M{"source": "builder"})

	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].msg != "builder test" {
		t.Errorf("msg = %q, want 'builder test'", logs[0].msg)
	}
	if logs[0].data["source"] != "builder" {
		t.Errorf("data['source'] = %v, want 'builder'", logs[0].data["source"])
	}
}
