package serverlog

import (
	"fmt"
	"strings"
)

// Printer 是日志输出边界，实现必须支持并发 Log，并负责自身资源的幂等关闭。
// Log 接收到的嵌套对象按只读处理；core 不持有配置锁调用输出器。
type Printer interface {
	// Log 接收一次输出的顶层字段副本，不保留可写共享 payload。
	Log(level Level, msg string, data M)
	// Close 释放输出资源，应幂等且等待已接受的异步任务。
	Close()
}

// LogFormat 将日志转换为不含行终止符的字符串，实现必须支持并发调用。
type LogFormat interface {
	// Format 返回不含换行终止符的字符串，不修改输入字段。
	Format(level Level, msg string, data M) (result string)
}

// PreHandler 转换单个输出器的数据，不应修改共享嵌套对象或重入持有它的关闭流程。
type PreHandler func(level Level, msg string, data M) M

////////////////////////////////////////////////////////////

// Level 表示日志等级，数值越小严重程度越高，范围从 PanicLevel 到 TraceLevel。
type Level uint32

// String 返回稳定的小写等级名称；无效值返回 unknown。
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

// ParseLevel parses a case-insensitive log level, ignoring surrounding whitespace.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid log level: %q", lvl)
}

// MarshalText 实现 encoding.TextMarshaler，无效等级返回错误。
func (level Level) MarshalText() ([]byte, error) {
	switch level {
	case TraceLevel:
		return []byte("trace"), nil
	case DebugLevel:
		return []byte("debug"), nil
	case InfoLevel:
		return []byte("info"), nil
	case WarnLevel:
		return []byte("warning"), nil
	case ErrorLevel:
		return []byte("error"), nil
	case FatalLevel:
		return []byte("fatal"), nil
	case PanicLevel:
		return []byte("panic"), nil
	}

	return nil, fmt.Errorf("not a valid log level %d", level)
}

// AllLevels 按严重程度从高到低列出内置等级；调用方应将该切片视为只读。
var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

// 等级按数值从小到大排列；设置最大等级表示允许该等级及更高严重程度输出。
const (
	// PanicLevel 是最高严重程度，输出后触发 panic。
	PanicLevel Level = iota
	// FatalLevel 输出并关闭资源后以状态码 1 退出，过滤不影响退出行为。
	FatalLevel
	// ErrorLevel 表示需要记录和诊断的错误，不改变程序控制流。
	ErrorLevel
	// WarnLevel 表示需要关注的异常情况，不改变程序控制流。
	WarnLevel
	// InfoLevel 表示正常运行信息。
	InfoLevel
	// DebugLevel 表示调试信息。
	DebugLevel
	// TraceLevel 表示更细粒度的诊断信息。
	TraceLevel
)
