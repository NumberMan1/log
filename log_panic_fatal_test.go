package serverlog

import (
	"sync"
	"testing"
)

func TestPanicLevel(t *testing.T) {
	l := NewLogger()
	l.SetMaxLevel(PanicLevel)

	l.AddPrinter(NewLevelSwitch().EnableLevels(PanicLevel), &mockPrinter{})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for PanicLevel")
		}
	}()
	l.Panic("panic test", nil)
}

func TestFatalLevel(t *testing.T) {
	l := NewLogger()
	l.SetMaxLevel(FatalLevel)

	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(FatalLevel), mp)

	exited := false
	exitCode := 0
	SetExitFunc(func(code int) {
		exited = true
		exitCode = code
	})

	l.Fatal("fatal test", nil)

	if !exited {
		t.Error("expected exit function to be called")
	}
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}
	// After Fatal recovers from exit (because we overrode it),
	// the mockPrinter should have the log
	logs := mp.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].msg != "fatal test" {
		t.Errorf("msg = %q, want 'fatal test'", logs[0].msg)
	}
}

func TestLogger_Panic_Recovers(t *testing.T) {
	l := NewLogger()
	mp := &mockPrinter{}
	l.AddPrinter(NewLevelSwitch().EnableLevels(PanicLevel), mp)

	panicked := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		l.Panic("panic in goroutine", nil)
	}()
	wg.Wait()

	if !panicked {
		t.Error("expected panic to be recovered")
	}
	if len(mp.GetLogs()) != 1 {
		t.Errorf("expected 1 log in printer, got %d", len(mp.GetLogs()))
	}
}

func TestSetExitFunc_Nil(t *testing.T) {
	SetExitFunc(nil)
	// Should not panic, and should reset to default
	fn := exitFn.Load().(func(int))
	if fn == nil {
		t.Error("exitFn should not be nil after SetExitFunc(nil)")
	}
	ResetExitFunc()
}

func TestResetExitFunc(t *testing.T) {
	SetExitFunc(func(code int) {})
	ResetExitFunc()

	// Cannot easily test that it calls os.Exit, but verify it's stored
	fn := exitFn.Load().(func(int))
	if fn == nil {
		t.Error("exitFn should not be nil after ResetExitFunc")
	}
}
