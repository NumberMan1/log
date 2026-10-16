package serverlog

import (
	"os"
	"sync"
)

// ConsolePrinterConfig 配置终端预处理、格式与调用位置解析。
type ConsolePrinterConfig struct {
	// PreHandler 可选，在格式化前处理独立顶层字段副本；实现必须并发安全。
	PreHandler PreHandler
	// LogFormat 为 nil 时采用文本格式：[等级] 时间 调用位置 消息 字段。
	LogFormat LogFormat
	// SkipFrame 在 ChangeSkipFrame 为 true 时指定文本格式器的栈跳过层数。
	SkipFrame int
	// ChangeSkipFrame 为 false 时自动查找库外调用位置，自定义 LogFormat 会忽略这两项。
	ChangeSkipFrame bool
}

// ConsolePrinter 按完整行写入标准输出；多个内置终端输出器不会交错写行。
type ConsolePrinter struct {
	preHandler PreHandler
	logFormat  LogFormat
}

// NewConsolePrinterDefault 创建自动解析调用位置的文本输出器。
func NewConsolePrinterDefault() *ConsolePrinter {
	return NewConsolePrinter(ConsolePrinterConfig{})
}

// NewConsolePrinter 创建终端输出器；自定义格式器和预处理函数必须并发安全。
func NewConsolePrinter(config ConsolePrinterConfig) *ConsolePrinter {
	printer := &ConsolePrinter{}

	if config.PreHandler == nil {
		printer.preHandler = func(level Level, msg string, data M) M { return data }
	} else {
		printer.preHandler = config.PreHandler
	}

	if config.LogFormat == nil {
		if config.ChangeSkipFrame {
			printer.logFormat = NewTextFormat(config.SkipFrame)
		} else {
			printer.logFormat = NewTextFormat(-1)
		}
	} else {
		printer.logFormat = config.LogFormat
	}

	return printer
}

var consoleMu sync.Mutex

// Log 格式化后写一整行；只复制顶层字段，嵌套对象按只读数据处理。
func (l *ConsolePrinter) Log(level Level, msg string, data M) {
	data = l.preHandler(level, msg, cloneData(data))

	logStr := l.logFormat.Format(level, msg, data)

	consoleMu.Lock()
	defer consoleMu.Unlock()
	_, _ = os.Stdout.WriteString(logStr + "\n")
}

// Close 尝试同步标准输出，但不关闭由进程拥有的标准输出句柄。
func (l *ConsolePrinter) Close() {
	_ = os.Stdout.Sync()
}
