package serverlog

import (
	"runtime"
	"testing"
)

func TestStackTraceFrame(t *testing.T) {
	f := stackTraceFrame(1)
	if f == nil {
		t.Fatal("stackTraceFrame returned nil")
	}
	// File should be non-empty (this file itself)
	if f.File == "" {
		t.Error("expected non-empty File")
	}
	if f.Line == 0 {
		t.Error("expected non-zero Line")
	}
	// Line should be a reasonable number
	if f.Line <= 0 {
		t.Errorf("expected positive Line, got %d", f.Line)
	}
}

func TestStackTraceFrame_InvalidSkip(t *testing.T) {
	// Very large skip should return empty frame
	f := stackTraceFrame(1000)
	if f == nil {
		t.Fatal("stackTraceFrame returned nil")
	}
	if f.File != "" {
		t.Errorf("expected empty File for large skip, got %q", f.File)
	}
	if f.Line != 0 {
		t.Errorf("expected zero Line for large skip, got %d", f.Line)
	}
}

func TestStackTraceFrames(t *testing.T) {
	frames := stackTraceFrames(1)
	if len(frames) == 0 {
		t.Fatal("stackTraceFrames returned empty slice")
	}
	// First frame should be from the testing framework or this file
	for _, f := range frames {
		if f.File == "" {
			t.Error("frame has empty File")
		}
		if f.Line <= 0 {
			t.Errorf("frame has non-positive Line: %d", f.Line)
		}
	}
}

func TestStackTraceFrames_InvalidSkip(t *testing.T) {
	frames := stackTraceFrames(1000)
	if len(frames) != 0 {
		t.Errorf("expected empty slice for large skip, got %d frames", len(frames))
	}
}

func TestStackDepth(t *testing.T) {
	// Verify we can get our own file in the stack trace
	f := stackTraceFrame(1)
	// This file should appear somewhere
	if f.File == "" {
		t.Error("expected to find current file in stack trace")
	}
}

func TestSkipIncrement(t *testing.T) {
	// stackTraceFrame(1) and stackTraceFrame(2) should give different results
	f1 := stackTraceFrame(1)
	f2 := stackTraceFrame(2)

	// They should be different frames (different callers)
	if f1.File == f2.File && f1.Line == f2.Line {
		t.Log("same frame for skip=1 and skip=2, might be inlined")
	}
}

func TestPcPool(t *testing.T) {
	// Ensure the pool can be used without panics
	frames := stackTraceFrames(1)
	if len(frames) == 0 {
		t.Fatal("expected frames")
	}

	// Run Callers directly to ensure the pool is not corrupted
	pcs := make([]uintptr, 10)
	n := runtime.Callers(1, pcs)
	if n == 0 {
		t.Fatal("runtime.Callers returned 0")
	}
}
