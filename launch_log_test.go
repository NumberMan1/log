package serverlog

import (
	"testing"
)

func TestLaunchLog_BootInfo(t *testing.T) {
	called := false
	var capturedArgs []any

	SetBootInfoPrinter(func(v ...any) {
		called = true
		capturedArgs = v
	})

	BootInfo("server starting", "on port", 8080)

	if !called {
		t.Error("BootInfo printer should have been called")
	}
	if capturedArgs == nil {
		t.Error("capturedArgs should not be nil")
	}

	// Restore
	SetBootInfoPrinter(func(v ...any) {})
}

func TestLaunchLog_GetBootInfoPrinter(t *testing.T) {
	fn := GetBootInfoPrinter()
	if fn == nil {
		t.Fatal("GetBootInfoPrinter returned nil")
	}
	// Should not panic
	fn("test message")
}

func TestLaunchLog_SetBootInfoPrinter(t *testing.T) {
	original := GetBootInfoPrinter()

	called := false
	SetBootInfoPrinter(func(v ...any) {
		called = true
	})

	BootInfo("test")
	if !called {
		t.Error("custom printer should have been called")
	}

	// Restore
	SetBootInfoPrinter(original)
}
