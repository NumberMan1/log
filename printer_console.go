package serverlog

import (
	"os"
)

type ConsolePrinterConfig struct {
	PreHandler      PreHandler
	LogFormat       LogFormat // 默认TextFormat，将以字符串格式打印，字符串格式: [INFO] Mon Jan 2 15:04:05 MST 2006 main.go:32 log data
	SkipFrame       int       // 获取文件和行数需要跳过几帧调用栈，用于校准 TextFormat 第三部分的 文件:行数
	ChangeSkipFrame bool      // SkipFrame如果有变更该值应该为true
}

type ConsolePrinter struct {
	preHandler PreHandler
	logFormat  LogFormat
}

func NewConsolePrinterDefault() *ConsolePrinter {
	return NewConsolePrinter(ConsolePrinterConfig{})
}

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
			printer.logFormat = NewTextFormat(6)
		}
	} else {
		printer.logFormat = config.LogFormat
	}

	return printer
}

var line = []byte("\n")

func (l *ConsolePrinter) Log(level Level, msg string, data M) {
	data = l.preHandler(level, msg, data)

	logStr := l.logFormat.Format(level, msg, data)

	os.Stdout.WriteString(logStr)
	os.Stdout.Write(line)
}

func (l *ConsolePrinter) Close() {
	_ = os.Stdout.Sync()
}
