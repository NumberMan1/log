package serverlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NumberMan1/log/field"
)

func TestConfiguredLoggerFileAndSentryRouting(t *testing.T) {
	dir := t.TempDir()
	tr := &captureSentryTransport{}
	raw, err := NewConfiguredLogger(Config{Fields: M{"component": "example"}, Level: "debug", LogFilePath: dir, SentryDSN: "https://public@example.invalid/1", Sentry: SentryPrinterConfig{Transport: tr}})
	if err != nil {
		t.Fatal(err)
	}
	logger := NewServerLogger(raw)
	logger.Debug("debug", field.String("detail", "value"))
	logger.Warn("warning")
	logger.Error("error", field.Error(os.ErrNotExist))
	raw.Close()
	lines := readLogLines(t, dir)
	if len(lines) != 3 {
		t.Fatalf("file records=%d", len(lines))
	}
	for _, line := range lines {
		var record struct {
			Data    map[string]any
			Message string
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record.Data["component"] != "example" || record.Data["caller"] == "" {
			t.Fatalf("missing fields: %s", line)
		}
		if record.Message == "error" && record.Data["error"] != os.ErrNotExist.Error() {
			t.Fatalf("nested error lost: %s", line)
		}
		for _, key := range []string{"app", "region", "server"} {
			if _, ok := record.Data[key]; ok {
				t.Fatalf("implicit business field %s", key)
			}
		}
	}
	if len(tr.snapshot()) != 1 {
		t.Fatal("only errors should reach Sentry")
	}
	for _, name := range []string{"log", "event"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected implicit directory %s: %v", name, err)
		}
	}
}

func TestConfigureFailureKeepsExistingDefault(t *testing.T) {
	old := GetDefaultLogger()
	if err := Configure(Config{Level: "nonsense"}); err == nil {
		t.Fatal("invalid config accepted")
	}
	if old != GetDefaultLogger() {
		t.Fatal("default changed after failure")
	}
	dir := t.TempDir()
	// Fail Sentry setup after file output has already opened; Configure must roll back.
	if err := Configure(Config{LogFilePath: dir, SentryDSN: "invalid"}); err == nil {
		t.Fatal("expected Sentry setup failure")
	}
	if old != GetDefaultLogger() {
		t.Fatal("partial initialization changed default")
	}
}

func TestIndependentDataLogger(t *testing.T) {
	old := GetDefaultLogger()
	defer func() { Close(); SetDefaultLogger(old) }()
	dir := t.TempDir()
	if err := Configure(Config{Fields: M{"label": "example"}, Level: "info", LogFilePath: filepath.Join(dir, "main")}); err != nil {
		t.Fatal(err)
	}
	DefaultLogger().Info("configured")
	data := NewLogger()
	printer, err := NewDataPrinter(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	data.AddPrinter(LevelSwitchLte(InfoLevel), printer)
	data.Info("", M{"kind": "sample"})
	data.Close()
	Close()
	if len(readLogLines(t, filepath.Join(dir, "main"))) != 1 {
		t.Fatal("main output missing")
	}
	lines := readLogLines(t, filepath.Join(dir, "data"))
	if len(lines) != 1 || !strings.Contains(lines[0], "sample") || !strings.Contains(lines[0], "Time") {
		t.Fatal(lines)
	}
}

func TestConfigFieldsAreSnapshotsAndCallFieldsOverride(t *testing.T) {
	defaults := M{"custom": "initial", "other": 1}
	raw, err := NewConfiguredLogger(Config{Fields: defaults})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	defaults["custom"] = "changed"
	p := &mockPrinter{}
	raw.AddPrinter(LevelSwitchLte(TraceLevel), p)
	raw.Info("first", nil)
	raw.Info("second", M{"custom": "call"})
	logs := p.GetLogs()
	if logs[0].data["custom"] != "initial" || logs[1].data["custom"] != "call" || logs[1].data["other"] != 1 {
		t.Fatalf("field merging: %#v", logs)
	}
}

func TestConfiguredLevelPrecedenceAndInvalidOutputs(t *testing.T) {
	t.Setenv("MAX_LOG_LEVEL", "warn")
	level, err := configuredLevel("")
	if err != nil || level != WarnLevel {
		t.Fatal(level, err)
	}
	level, err = configuredLevel(" DEBUG ")
	if err != nil || level != DebugLevel {
		t.Fatal(level, err)
	}
	if _, err := NewConfiguredLogger(Config{StdoutTyp: "unknown"}); err == nil {
		t.Fatal("invalid stdout accepted")
	}
	if _, err := NewConfiguredLogger(Config{SentryDSN: "invalid"}); err == nil {
		t.Fatal("invalid Sentry DSN accepted")
	}
}

func TestBuilderFileAndSentryFactories(t *testing.T) {
	tr := &captureSentryTransport{}
	dir := t.TempDir()
	raw := NewBuilder().AddFilePrinter(InfoLevel, FilePrinterConfig{Dir: filepath.Join(dir, "data")}).
		AddJsonlogPrinter(InfoLevel, filepath.Join(dir, "json")).
		AddConsolePrinter(InfoLevel, ConsolePrinterConfig{}).
		AddSentryPrinter(ErrorLevel, SentryPrinterConfig{Dsn: "https://public@example.invalid/1", Transport: tr}).Logger()
	raw.Error("builder", M{"value": 1})
	raw.Close()
	if len(readLogLines(t, dir)) != 2 || len(tr.snapshot()) != 1 {
		t.Fatal("builder outputs missing")
	}
}

func TestLegacyInitOnlyInitializesOnce(t *testing.T) {
	old := GetDefaultLogger()
	initMu.Lock()
	wasInited := inited
	inited = false
	initMu.Unlock()
	defer func() { Close(); SetDefaultLogger(old); initMu.Lock(); inited = wasInited; initMu.Unlock() }()
	Init(t.TempDir(), "")
	first := GetDefaultLogger()
	Init("", "")
	if GetDefaultLogger() != first {
		t.Fatal("legacy Init should only initialize once")
	}
}
