package serverlog

import "github.com/NumberMan1/log/field"

// Logger 提供带字段的日志接口；With 和 Clone 派生独立 scope，不拥有底层 core。
type Logger interface {
	// With 追加字段并返回新 scope，后追加的同名字段优先。
	With(...field.Field) Logger
	// Clone 复制字段 scope，共享原有输出 core。
	Clone() Logger
	// Debug 记录调试日志，本次字段覆盖 scope 中的同名字段。
	Debug(string, ...field.Field)
	// Info 记录普通运行信息。
	Info(string, ...field.Field)
	// Warn 记录需要关注的情况，不改变程序控制流。
	Warn(string, ...field.Field)
	// Error 记录错误，保留原始 error 供输出器处理。
	Error(string, ...field.Field)
}

// ServerLogger 将字段 scope 适配到 L；并发打印和派生子 scope 安全。
type ServerLogger struct {
	logger *L
	fields field.Fields
}

// NewServerLogger 包装指定 core；nil 时创建一个没有输出器的新 core。
// 返回值不负责关闭 core，调用方仍需管理其生命周期。
func NewServerLogger(raw *L) Logger {
	if raw == nil {
		raw = NewLogger()
	}
	return &ServerLogger{logger: raw}
}

// DefaultLogger 为当前默认 core 创建字段 logger；取得的引用不会跟随后续替换。
func DefaultLogger() Logger { return NewServerLogger(GetDefaultLogger()) }

// NewLoggerFromParent 复制父 scope 并追加字段；parent 为 nil 时使用默认 logger。
func NewLoggerFromParent(parent Logger, fields ...field.Field) Logger {
	if parent == nil {
		parent = DefaultLogger()
	}
	return parent.Clone().With(fields...)
}

// With 返回追加字段后的新 scope，不修改父 scope；同名字段以后追加的值为准。
func (logger *L) With(fields ...field.Field) Logger { return NewServerLogger(logger).With(fields...) }

// With 返回追加字段后的新 scope，不修改父 scope；同名字段以后追加的值为准。
func (logger *ServerLogger) With(fields ...field.Field) Logger {
	return &ServerLogger{logger: logger.logger, fields: logger.fields.With(fields...)}
}

// Clone 复制字段切片并共享输出 core；嵌套字段值仍应视为只读。
func (logger *ServerLogger) Clone() Logger { return logger.With() }

// Debug 合并 scope 和本次调用的字段，记录 Debug 日志。
func (logger *ServerLogger) Debug(msg string, fields ...field.Field) {
	logger.log(DebugLevel, msg, fields)
}

// Info 合并 scope 和本次调用的字段，记录 Info 日志。
func (logger *ServerLogger) Info(msg string, fields ...field.Field) {
	logger.log(InfoLevel, msg, fields)
}

// Warn 合并 scope 和本次调用的字段，记录 Warn 日志。
func (logger *ServerLogger) Warn(msg string, fields ...field.Field) {
	logger.log(WarnLevel, msg, fields)
}

// Error 合并 scope 和本次调用的字段，保留原始 error 供输出器处理。
func (logger *ServerLogger) Error(msg string, fields ...field.Field) {
	logger.log(ErrorLevel, msg, fields)
}
func (logger *ServerLogger) log(level Level, msg string, fields []field.Field) {
	logger.logger.Log(level, msg, logger.fields.With(fields...).ToMap())
}
