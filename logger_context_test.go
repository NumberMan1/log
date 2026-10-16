package serverlog

import (
	"context"
	"errors"
	"github.com/NumberMan1/log/field"
	"sync"
	"testing"
)

func TestContextScopesRemainIndependent(t *testing.T) {
	raw := NewLogger()
	p := &mockPrinter{}
	raw.AddPrinter(LevelSwitchLte(TraceLevel), p)
	root := raw.With(field.String("service", "api"))
	ctx := WithContext(context.Background(), root)
	a := WithFields(ctx, field.String("request_id", "a"))
	b := WithFields(ctx, field.String("request_id", "b"))
	MustFromContext(a).Info("a", field.Int("status", 200))
	MustFromContext(b).Warn("b")
	MustFromContext(ctx).Error("root", field.Error(errors.New("original")))
	logs := p.GetLogs()
	if logs[0].data["request_id"] != "a" || logs[1].data["request_id"] != "b" {
		t.Fatal("sibling scopes mixed")
	}
	if _, ok := logs[2].data["request_id"]; ok {
		t.Fatal("child mutated parent")
	}
	if _, ok := logs[2].data["error"].(error); !ok {
		t.Fatal("error type lost")
	}
	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("unexpected logger")
	}
	if _, err := LoggerFromContext(context.Background()); !errors.Is(err, ErrNoLoggerFound) {
		t.Fatal(err)
	}
	if MustFromContext(nil) == nil {
		t.Fatal("missing fallback")
	}
}

func TestConcurrentSiblingScopes(t *testing.T) {
	raw := NewLogger()
	p := &mockPrinter{}
	raw.AddPrinter(LevelSwitchLte(InfoLevel), p)
	parent := raw.With(field.String("a", "a"), field.String("b", "b"), field.String("c", "c"))
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			NewLoggerFromParent(parent, field.Int("child", i)).Info("child", field.Int("child", i+1))
		}(i)
	}
	wg.Wait()
	for _, entry := range p.GetLogs() {
		if entry.data["child"].(int) < 1 {
			t.Fatal("fields lost")
		}
	}
	parent.Info("parent")
	if _, ok := p.GetLogs()[100].data["child"]; ok {
		t.Fatal("scope leaked")
	}
}
