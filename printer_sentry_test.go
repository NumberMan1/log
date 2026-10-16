package serverlog

import (
	"context"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

type captureSentryTransport struct {
	mu              sync.Mutex
	events          []*sentry.Event
	flushes, closes int
	timeout         bool
}

func (*captureSentryTransport) Configure(sentry.ClientOptions) {}
func (tr *captureSentryTransport) SendEvent(event *sentry.Event) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.events = append(tr.events, event)
}
func (tr *captureSentryTransport) Flush(time.Duration) bool {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.flushes++
	return !tr.timeout
}
func (tr *captureSentryTransport) FlushWithContext(context.Context) bool { return tr.Flush(0) }
func (tr *captureSentryTransport) Close()                                { tr.mu.Lock(); defer tr.mu.Unlock(); tr.closes++ }
func (tr *captureSentryTransport) snapshot() []*sentry.Event {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return append([]*sentry.Event(nil), tr.events...)
}

func TestSentryNativeExceptionAndStructuredPayload(t *testing.T) {
	tr := &captureSentryTransport{}
	p := NewSentryPrinter(SentryPrinterConfig{
		Dsn: "https://public@example.invalid/1", Environment: "test", Release: "v1", Transport: tr,
		UserIDKeys:  []string{"identity", "fallback_identity"},
		TagKeys:     []string{"category"},
		ExcludeKeys: []string{"local_debug"},
	})
	raw := NewLogger()
	raw.AddPrinter(LevelSwitchLte(ErrorLevel), p)
	original := fmt.Errorf("query failed: %w", errors.New("database unavailable"))
	request := M{"action": "read", "ids": []int{1, 2}}
	data := M{"error": original, "input": request, "output": "", "identity": 0, "fallback_identity": 42, "category": "io", "local_debug": "local only"}
	raw.Error("request failed", data)
	raw.Info("not an error", nil)
	raw.Close()
	events := tr.snapshot()
	if len(events) != 1 {
		t.Fatalf("events=%d", len(events))
	}
	event := events[0]
	if event.Level != sentry.LevelError || event.Message != "request failed" || event.Environment != "test" || event.Release != "v1" {
		t.Fatalf("event metadata: %#v", event)
	}
	if len(event.Exception) != 2 {
		t.Fatalf("unwrap chain lost: %#v", event.Exception)
	}
	for _, exception := range event.Exception {
		if exception.Stacktrace == nil || len(exception.Stacktrace.Frames) == 0 {
			t.Fatal("missing exception stack")
		}
	}
	if event.User.ID != "42" || event.Tags["category"] != "io" {
		t.Fatalf("wrong user/tags: %#v", event)
	}
	extra := event.Contexts["extra"]
	for _, key := range []string{"error", "fallback_identity", "category", "local_debug"} {
		if _, ok := extra[key]; ok {
			t.Fatalf("unexpected extra key %s", key)
		}
	}
	payload, ok := extra["input"].(M)
	if !ok || payload["action"] != "read" {
		t.Fatalf("payload is not structured: %#v", extra)
	}
	if _, ok := event.Tags["input"]; ok {
		t.Fatal("payload used as tag")
	}
	if data["error"] != original || request["action"] != "read" {
		t.Fatal("input modified")
	}
	if tr.flushes != 1 || tr.closes != 1 {
		t.Fatalf("lifecycle: flush=%d close=%d", tr.flushes, tr.closes)
	}
}

func TestSentryCustomMapperAndNoGlobalScopeLeak(t *testing.T) {
	sentry.ConfigureScope(func(scope *sentry.Scope) { scope.SetTag("global", "must not leak") })
	defer sentry.ConfigureScope(func(scope *sentry.Scope) { scope.RemoveTag("global") })
	tr := &captureSentryTransport{}
	p := NewSentryPrinter(SentryPrinterConfig{Dsn: "https://public@example.invalid/1", Transport: tr,
		FieldMapper: SentryFieldMapperFunc(func(event SentryLogEvent) SentryFields {
			return SentryFields{User: sentry.User{ID: "custom"}, Tags: map[string]string{"mapped": event.Message}, Extra: M{"payload": event.Data["payload"]}}
		}), ExtraTagFunc: func(tags map[string]string) { tags["extra"] = "tag" },
	})
	defer p.Close()
	p.Log(WarnLevel, "mapped", M{"payload": M{"ok": true}})
	event := tr.snapshot()[0]
	if event.User.ID != "custom" || event.Level != sentry.LevelWarning || event.Tags["mapped"] != "mapped" || event.Tags["extra"] != "tag" {
		t.Fatal("custom mapper not applied")
	}
	if _, ok := event.Tags["global"]; ok {
		t.Fatal("global scope leaked")
	}
}

func TestSentryNormalizationBoundsAndNoMutation(t *testing.T) {
	mapper := NewDefaultSentryFieldMapper(SentryPrinterConfig{MaxFieldLength: 8, TagKeys: []string{}})
	cycle := M{}
	cycle["self"] = cycle
	nested := M{"text": "1234567界", "number": 12, "list": []string{"1234567界"}}
	fields := mapper.Map(SentryLogEvent{Data: M{"input": nested, "cycle": cycle, "long": strings.Repeat("界", 100), "status": 500, "error": "plain text"}})
	if len(fields.Tags) != 0 {
		t.Fatal("empty tag list should disable tags")
	}
	if fields.Extra["error"] != "plain te" {
		t.Fatal("string error was not preserved as bounded data")
	}
	text := fields.Extra["long"].(string)
	if len(text) > 8 || !utf8.ValidString(text) {
		t.Fatalf("invalid utf8: %q", text)
	}
	result := fields.Extra["input"].(M)
	if result["number"] != 12 || result["text"] != "1234567" || nested["text"] != "1234567界" {
		t.Fatalf("normalization changed input: %#v", result)
	}
	if fields.Extra["cycle"] == nil {
		t.Fatal("cyclic data lost")
	}
}

func TestSentryFailureIsObservableAndDoesNotExit(t *testing.T) {
	tr := &captureSentryTransport{timeout: true}
	var failures []error
	p := NewSentryPrinter(SentryPrinterConfig{Dsn: "https://public@example.invalid/1", Transport: tr,
		OnError:     func(err error) { failures = append(failures, err) },
		FieldMapper: SentryFieldMapperFunc(func(SentryLogEvent) SentryFields { panic("broken mapper") }),
	})
	p.Log(ErrorLevel, "failure", nil)
	if p.Err() == nil || len(failures) != 1 {
		t.Fatal("mapper failure not reported")
	}
	p.Close()
	p.Close()
	p.Log(ErrorLevel, "after close", nil)
	if len(failures) != 2 || !strings.Contains(p.Err().Error(), "timed out") {
		t.Fatalf("flush failure missing: %v", failures)
	}
}

func TestSentryConcurrentCloseAndLog(t *testing.T) {
	tr := &captureSentryTransport{}
	p := NewSentryPrinter(SentryPrinterConfig{Dsn: "https://public@example.invalid/1", Transport: tr})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				p.Log(ErrorLevel, "concurrent", M{"n": j})
			}
			p.Close()
		}()
	}
	wg.Wait()
	count := len(tr.snapshot())
	p.Log(ErrorLevel, "late", nil)
	if len(tr.snapshot()) != count || tr.closes != 1 || tr.flushes != 1 {
		t.Fatal("close lifecycle broken")
	}
}

func TestSentryInvalidConfigurationReturnsError(t *testing.T) {
	if _, err := NewSentryPrinterWithError(SentryPrinterConfig{Dsn: "invalid"}); err == nil {
		t.Fatal("expected invalid DSN error")
	}
}

func TestSentryDefaultMappingDoesNotInterpretApplicationFields(t *testing.T) {
	source := M{"app": "sample", "region": "test", "pid": 12, "user_id": "abc", "request": nil, "response": "", "error_stack": "keep me", "log_kind": "custom"}
	fields := NewDefaultSentryFieldMapper(SentryPrinterConfig{}).Map(SentryLogEvent{Data: source})
	if len(fields.Tags) != 0 || fields.User.ID != "" {
		t.Fatalf("implicit tag/user mapping: %#v", fields)
	}
	if len(fields.Extra) != len(source) {
		t.Fatalf("application fields removed: %#v", fields.Extra)
	}
	for key, value := range source {
		if fields.Extra[key] != value {
			t.Fatalf("field %s changed: %#v", key, fields.Extra[key])
		}
	}
}

func TestSentryIgnoresApplicationEnvironmentVariables(t *testing.T) {
	t.Setenv("SERVER_TYPE", "custom-application-environment")
	t.Setenv("SERVER_NAME", "custom-application-host")
	t.Setenv("SENTRY_ENVIRONMENT", "")
	tr := &captureSentryTransport{}
	raw := NewBuilder().AddSentryPrinter(ErrorLevel, SentryPrinterConfig{Dsn: "https://public@example.invalid/1", Transport: tr}).Logger()
	raw.Error("sample", nil)
	raw.Close()
	event := tr.snapshot()[0]
	if event.Environment == "custom-application-environment" || event.ServerName == "custom-application-host" {
		t.Fatal("application environment variables affected generic logger")
	}
}
