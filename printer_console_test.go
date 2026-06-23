package serverlog

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestConsolePrinter_Log_DefaultFormat(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cp := NewConsolePrinterDefault()
	cp.Log(InfoLevel, "test console", M{"key": "value"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Errorf("expected 'INFO' in output, got: %s", output)
	}
	if !strings.Contains(output, "test console") {
		t.Errorf("expected 'test console' in output, got: %s", output)
	}
}

func TestConsolePrinter_Log_CustomFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cp := NewConsolePrinter(ConsolePrinterConfig{
		LogFormat: NewOnlyMessageFormat(),
	})
	cp.Log(InfoLevel, "custom format test", nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should only contain the message
	if !strings.Contains(output, "custom format test") {
		t.Errorf("expected 'custom format test' in output, got: %s", output)
	}
	if strings.Contains(output, "INFO") {
		t.Errorf("should not contain level in only-message format, got: %s", output)
	}
}

func TestConsolePrinter_Log_WithPreHandler(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cp := NewConsolePrinter(ConsolePrinterConfig{
		PreHandler: StackWrapPreHandler,
		LogFormat:  NewJsonFormat(),
	})
	cp.Log(InfoLevel, "prehandler test", M{"a": 1})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain JSON with wrapped fields
	if !strings.Contains(output, "LogType") {
		t.Errorf("expected 'LogType' in output, got: %s", output)
	}
	if !strings.Contains(output, "Message") {
		t.Errorf("expected 'Message' in output, got: %s", output)
	}
}

func TestConsolePrinter_Log_ChangeSkipFrame(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cp := NewConsolePrinter(ConsolePrinterConfig{
		ChangeSkipFrame: true,
		SkipFrame:       4,
	})
	cp.Log(InfoLevel, "skip test", nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should still produce valid output
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestConsolePrinter_Close(t *testing.T) {
	cp := NewConsolePrinterDefault()
	// Should not panic
	cp.Close()
}
