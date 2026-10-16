package serverlog_test

import (
	"context"
	"fmt"
	serverlog "github.com/NumberMan1/log"
	"github.com/NumberMan1/log/field"
)

type examplePrinter struct{}

func (examplePrinter) Log(level serverlog.Level, msg string, data serverlog.M) {
	fmt.Println(level, msg, data["service"], data["request_id"])
}
func (examplePrinter) Close() {}

func ExampleWithFields() {
	raw := serverlog.NewLogger()
	raw.AddPrinter(serverlog.LevelSwitchLte(serverlog.InfoLevel), examplePrinter{})
	defer raw.Close()
	logger := raw.With(field.String("service", "api"))
	ctx := serverlog.WithContext(context.Background(), logger)
	ctx = serverlog.WithFields(ctx, field.String("request_id", "req-123"))
	serverlog.MustFromContext(ctx).Info("started")
	// Output: info started api req-123
}
