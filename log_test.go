package serverlog

import (
	"sync"
	"testing"
)

// mockPrinter implements Printer for testing
type mockPrinter struct {
	logs      []logEntry
	closed    bool
	closeOnce sync.Once
	mu        sync.Mutex
}

type logEntry struct {
	level Level
	msg   string
	data  M
}

func (m *mockPrinter) Log(level Level, msg string, data M) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, logEntry{level, msg, data})
}

func (m *mockPrinter) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

func (m *mockPrinter) GetLogs() []logEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]logEntry, len(m.logs))
	copy(result, m.logs)
	return result
}

func (m *mockPrinter) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	if l == nil {
		t.Fatal("NewLogger returned nil")
	}
	if l.levelPrinters == nil {
		t.Error("levelPrinters should not be nil")
	}
	if l.maxLevel != TraceLevel {
		t.Errorf("maxLevel = %d, want %d", l.maxLevel, TraceLevel)
	}
}

func TestLogger_AddPrinter_RemovePrinter(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	ls := NewLevelSwitch().EnableLevels(InfoLevel)

	l.AddPrinter(ls, mp)
	if l.GetLevelSwitcherByPrinter(mp) != ls {
		t.Error("GetLevelSwitcherByPrinter should return the level switch")
	}

	l.RemovePrinter(mp)
	if l.GetLevelSwitcherByPrinter(mp) != nil {
		t.Error("GetLevelSwitcherByPrinter should return nil after removal")
	}
}

func TestLogger_SetPreHandler(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	called := false
	l.SetPreHandler(func(level Level, msg string, data M) (Level, string, M) {
		called = true
		return level, "modified: " + msg, data
	})

	l.Log(InfoLevel, "hello", nil)
	logs := mp.GetLogs()

	if !called {
		t.Error("preHandler should have been called")
	}
	if len(logs) == 0 {
		t.Fatal("expected at least one log entry")
	}
	if logs[0].msg != "modified: hello" {
		t.Errorf("msg = %q, want 'modified: hello'", logs[0].msg)
	}
}

func TestLogger_Debug(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(DebugLevel), mp)

	l.Debug("debug msg", M{"k": "v"})
	logs := mp.GetLogs()

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != DebugLevel {
		t.Errorf("level = %d, want %d", logs[0].level, DebugLevel)
	}
	if logs[0].msg != "debug msg" {
		t.Errorf("msg = %q, want 'debug msg'", logs[0].msg)
	}
	if logs[0].data["k"] != "v" {
		t.Errorf("data['k'] = %v, want 'v'", logs[0].data["k"])
	}
}

func TestLogger_Info(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	l.Info("info msg", nil)
	logs := mp.GetLogs()

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != InfoLevel {
		t.Errorf("level = %d, want %d", logs[0].level, InfoLevel)
	}
}

func TestLogger_Error(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(ErrorLevel), mp)

	l.Error("error msg", M{"err": "something"})
	logs := mp.GetLogs()

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != ErrorLevel {
		t.Errorf("level = %d, want %d", logs[0].level, ErrorLevel)
	}
}

func TestLogger_MultiplePrinters(t *testing.T) {
	l := NewLogger()
	mp1 := &mockPrinter{}
	mp2 := &mockPrinter{}

	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp1)
	l.AddPrinter(NewLevelSwitch().EnableLevels(ErrorLevel), mp2)

	l.Info("info only", nil)
	if len(mp1.GetLogs()) != 1 {
		t.Errorf("mp1 should have 1 info log, got %d", len(mp1.GetLogs()))
	}
	if len(mp2.GetLogs()) != 0 {
		t.Error("mp2 should not have info log")
	}

	l.Error("error msg", nil)
	if len(mp1.GetLogs()) != 1 {
		t.Errorf("mp1 should still have 1 log (only InfoLevel enabled), got %d", len(mp1.GetLogs()))
	}
	if len(mp2.GetLogs()) != 1 {
		t.Errorf("mp2 should have 1 error log, got %d", len(mp2.GetLogs()))
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}

	// Only enable Error level and below
	l.AddPrinter(LevelSwitchLte(ErrorLevel), mp)

	l.Debug("debug", nil)
	l.Info("info", nil)
	l.Log(WarnLevel, "warn", nil)
	l.Error("error", nil)

	logs := mp.GetLogs()
	// Debug, Info, Warn should be filtered out
	if len(logs) != 1 {
		t.Fatalf("expected 1 log (error only), got %d", len(logs))
	}
	if logs[0].level != ErrorLevel {
		t.Errorf("expected ErrorLevel, got %d", logs[0].level)
	}
}

func TestLogger_SetMaxLevel(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(AllLevels...), mp)

	l.SetMaxLevel(InfoLevel)

	l.Debug("debug", nil)
	l.Log(TraceLevel, "trace", nil)
	l.Info("info", nil)

	logs := mp.GetLogs()
	// Debug and Trace > InfoLevel, so they should be filtered
	if len(logs) != 1 {
		t.Fatalf("expected 1 log (info only), got %d", len(logs))
	}
	if logs[0].level != InfoLevel {
		t.Errorf("expected InfoLevel, got %d", logs[0].level)
	}
}

func TestLogger_Close(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(InfoLevel), mp)

	l.Close()
	if !mp.IsClosed() {
		t.Error("printer should be closed")
	}
}

func TestLogger_Trace(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(TraceLevel), mp)

	l.Log(TraceLevel, "trace msg", nil)
	logs := mp.GetLogs()

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != TraceLevel {
		t.Errorf("level = %d, want %d", logs[0].level, TraceLevel)
	}
}

func TestLogger_Warn(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(WarnLevel), mp)

	l.Log(WarnLevel, "warn msg", nil)
	logs := mp.GetLogs()

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].level != WarnLevel {
		t.Errorf("level = %d, want %d", logs[0].level, WarnLevel)
	}
}
