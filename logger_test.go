package log

import (
	"context"
	"testing"

	"github.com/NumberMan1/general/log"
	"github.com/NumberMan1/log/field"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDefaultEncoder(t *testing.T) {
	tests := []struct {
		name string
		want zapcore.EncoderConfig
	}{
		{name: "default", want: DefaultEncoder()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultEncoder(); &got == nil {
				t.Errorf("DefaultEncoder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	tests := []struct {
		name string
		want log.Logger
	}{
		{
			"logger",
			GetLogger(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLogger(); got == nil {
				t.Errorf("GetLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLoggerCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want log.Logger
	}{
		{name: "loggerCtx", args: args{ctx: context.Background()}, want: GetLoggerCtx(context.Background())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLoggerCtx(tt.args.ctx); got != nil {
				got.Info("info log", field.String("string", "ok"))
			}
		})
	}
}

func TestInit(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "init",
			args: args{opts: []Option{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.args.opts...)
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	t.Run("debug", func(t *testing.T) {
		l := New()
		l.Debug("msg", field.String("msg", "ok"))
	})
}

func TestLogger_Error(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		l := New()
		l.Error("msg", field.String("msg", "ok"))
	})
}

func TestLogger_Info(t *testing.T) {
	t.Run("info", func(t *testing.T) {
		l := New()
		l.Info("info", field.String("msg", "ok"))
	})
}

func TestLogger_Sync(t *testing.T) {
	t.Run("logger", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		l := &Logger{
			base: logger,
			fs:   []field.Field{},
		}
		l.Sync()
	})
}

func TestLogger_Warn(t *testing.T) {
	t.Run("warn", func(t *testing.T) {
		l := New()
		l.Warn("warn", field.String("msg", "ok"))
	})
}
