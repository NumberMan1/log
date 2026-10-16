package serverlog

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDisableLevelIsIdempotent(t *testing.T) {
	ls := NewLevelSwitch()
	for i := 0; i < 3; i++ {
		ls.DisableSingleLevel(InfoLevel)
		if ls.IsEnabled(InfoLevel) {
			t.Fatal("disabled level was enabled")
		}
	}
}

func TestConcurrentConfigurationAndLogging(t *testing.T) {
	l := NewLogger()
	p := &mockPrinter{}
	ls := LevelSwitchLte(TraceLevel)
	l.AddPrinter(ls, p)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				l.SetMaxLevel(DebugLevel)
				l.SetPreHandler(func(level Level, msg string, data M) (Level, string, M) { return level, msg, data })
				l.AddPrinter(ls, p)
				l.Info("concurrent", M{"error": errors.New("original")})
				l.GetLevelSwitcherByPrinter(p)
				l.RemovePrinter(p)
				ls.EnableSingleLevel(InfoLevel).DisableSingleLevel(TraceLevel).EnableLevelLte(DebugLevel)
			}
		}()
	}
	wg.Wait()
	l.Close()
}

func TestConcurrentEnablesDoNotLoseUpdates(t *testing.T) {
	for trial := 0; trial < 100; trial++ {
		ls := NewLevelSwitch()
		var wg sync.WaitGroup
		for _, level := range AllLevels {
			wg.Add(1)
			go func(level Level) { defer wg.Done(); ls.EnableSingleLevel(level) }(level)
		}
		wg.Wait()
		for _, level := range AllLevels {
			if !ls.IsEnabled(level) {
				t.Fatalf("lost level %v", level)
			}
		}
	}
}

type modifyingPrinter struct{}

func (*modifyingPrinter) Log(_ Level, _ string, data M) {
	data["error"] = "changed"
	data["extra"] = true
}
func (*modifyingPrinter) Close() {}

func TestErrorAndPrinterInputIsolation(t *testing.T) {
	l := NewLogger()
	capture := &mockPrinter{}
	l.AddPrinter(LevelSwitchLte(TraceLevel), &modifyingPrinter{})
	l.AddPrinter(LevelSwitchLte(TraceLevel), capture)
	err := errors.New("original")
	data := M{"error": err}
	for i := 0; i < 50; i++ {
		l.Error("failure", data)
	}
	if data["error"] != err || len(data) != 1 {
		t.Fatalf("caller data changed: %#v", data)
	}
	for _, entry := range capture.GetLogs() {
		if entry.data["error"] != err || len(entry.data) != 1 {
			t.Fatal("printers share mutable top-level fields")
		}
	}
	var formatted map[string]any
	if e := json.Unmarshal([]byte(NewJsonFormat().Format(ErrorLevel, "failure", data)), &formatted); e != nil {
		t.Fatal(e)
	}
	if formatted["error"] != "original" || data["error"] != err {
		t.Fatal("format must stringify error without changing source")
	}
}

func TestJSONFallbackIsValid(t *testing.T) {
	result := NewJsonFormat().Format(InfoLevel, "quoted \"message\"\n", M{"bad": make(chan int), "text": "\"\\\n"})
	if !json.Valid([]byte(result)) {
		t.Fatalf("invalid fallback JSON: %s", result)
	}
}

func TestFatalSurvivesFiltering(t *testing.T) {
	defer ResetExitFunc()
	old := GetDefaultLogger()
	SetDefaultLogger(NewLogger())
	defer SetDefaultLogger(old)
	l := NewLogger()
	l.SetMaxLevel(PanicLevel)
	called := false
	SetExitFunc(func(code int) { called = code == 1 })
	l.Fatal("must exit", nil)
	if !called {
		t.Fatal("Fatal skipped exit when filtered")
	}
}

type blockingPrinter struct {
	entered chan struct{}
	release chan struct{}
	closes  atomic.Int32
}

func (p *blockingPrinter) Log(Level, string, M) { close(p.entered); <-p.release }
func (p *blockingPrinter) Close()               { p.closes.Add(1) }

func TestCloseWaitsForInFlightLog(t *testing.T) {
	l := NewLogger()
	p := &blockingPrinter{entered: make(chan struct{}), release: make(chan struct{})}
	l.AddPrinter(LevelSwitchLte(InfoLevel), p)
	go l.Info("in flight", nil)
	<-p.entered
	done := make(chan struct{})
	go func() { l.Close(); close(done) }()
	select {
	case <-done:
		t.Fatal("Close returned before in-flight log")
	case <-time.After(10 * time.Millisecond):
	}
	close(p.release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Close stuck")
	}
	l.Close()
	if p.closes.Load() != 1 {
		t.Fatal("printer closed multiple times")
	}
	l.Info("after close", nil)
}

func TestDefaultReplacementConcurrentWithLogging(t *testing.T) {
	old := GetDefaultLogger()
	defer SetDefaultLogger(old)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				SetDefaultLogger(NewLogger())
				Info("message", nil)
				GetLevelSwitcherByPrinter(&mockPrinter{})
			}
		}()
	}
	wg.Wait()
}

func TestJSONHandlesCyclicAndWrappedErrors(t *testing.T) {
	cyclic := M{}
	cyclic["self"] = cyclic
	data := M{"cycle": cyclic, "Data": M{"error": errors.New("original")}}
	result := NewJsonFormat().Format(ErrorLevel, "failed", data)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["Data"].(map[string]any)["error"] != "original" {
		t.Fatal("nested error lost")
	}
}

func TestBootPrinterConcurrentReplacement(t *testing.T) {
	old := GetBootInfoPrinter()
	defer SetBootInfoPrinter(old)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				SetBootInfoPrinter(func(...any) {})
				BootInfo("concurrent")
			}
		}()
	}
	wg.Wait()
	SetBootInfoPrinter(nil)
	BootInfo("disabled")
}
