package serverlog

import (
	"context"
	"errors"
	"github.com/NumberMan1/log/field"
)

type contextKey struct{}

var loggerKey contextKey

// ErrNoLoggerFound 表示 context 中没有绑定 Logger。
var ErrNoLoggerFound = errors.New("no logger found")

// WithContext 返回绑定 logger 的标准派生 context，调用方必须保留返回值。
// logger 为 nil 时使用当前默认 logger；ctx 遵循 context.WithValue 的非 nil 约定。
func WithContext(ctx context.Context, logger Logger) context.Context {
	if logger == nil {
		logger = DefaultLogger()
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext 读取绑定的 Logger；ctx 为 nil 或缺失 logger 时返回 nil, false。
func FromContext(ctx context.Context) (Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	logger, ok := ctx.Value(loggerKey).(Logger)
	return logger, ok && logger != nil
}

// MustFromContext 优先读取 context，缺失时回退到当前默认 logger，不触发 panic。
func MustFromContext(ctx context.Context) Logger {
	if logger, ok := FromContext(ctx); ok {
		return logger
	}
	return DefaultLogger()
}

// LoggerFromContext 读取绑定的 Logger，缺失时返回 ErrNoLoggerFound。
func LoggerFromContext(ctx context.Context) (Logger, error) {
	if logger, ok := FromContext(ctx); ok {
		return logger, nil
	}
	return nil, ErrNoLoggerFound
}

// MustLoggerFromContext 是 MustFromContext 的兼容别名。
func MustLoggerFromContext(ctx context.Context) Logger { return MustFromContext(ctx) }

// WithFields 返回带新增字段的派生 context，父 context 和父 scope 保持不变。
// 父 context 没有 logger 时从当前默认 logger 派生。
func WithFields(ctx context.Context, fields ...field.Field) context.Context {
	return WithContext(ctx, MustFromContext(ctx).With(fields...))
}
