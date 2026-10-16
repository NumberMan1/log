package serverlog

import "fmt"

// CompatibleLogger 将 L 适配为 io.Writer、Printf 和字符串等级日志接口。
type CompatibleLogger struct {
	l *L
}

// Write 以 Info 等级记录字节内容，成功时返回 len(p), nil；保留原始换行。
func (i *CompatibleLogger) Write(p []byte) (n int, err error) {
	i.l.Log(InfoLevel, string(p), nil)
	return len(p), nil
}

// Printf 使用 fmt.Sprintf 格式化消息后以 Info 等级记录。
func (i *CompatibleLogger) Printf(format string, args ...interface{}) {
	i.l.Log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Log 解析字符串等级并分发；无法识别的等级回退到 Error。
func (i *CompatibleLogger) Log(levelStr string, message string, data map[string]interface{}) {
	level, err := ParseLevel(levelStr)
	if err != nil {
		level = ErrorLevel
	}
	i.l.Log(level, message, data)
}

// NewCompatibleLogger 包装非 nil core，所有写入和关闭责任仍属于原 core。
func NewCompatibleLogger(l *L) *CompatibleLogger {
	return &CompatibleLogger{l}
}
