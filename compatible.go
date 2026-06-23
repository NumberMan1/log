package serverlog

import "fmt"

type CompatibleLogger struct {
	l *L
}

func (i *CompatibleLogger) Write(p []byte) (n int, err error) {
	i.l.Log(InfoLevel, string(p), nil)
	return
}

func (i *CompatibleLogger) Printf(format string, args ...interface{}) {
	i.l.Log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

func (i *CompatibleLogger) Log(levelStr string, message string, data map[string]interface{}) {
	level, err := ParseLevel(levelStr)
	if err != nil {
		level = ErrorLevel
	}
	i.l.Log(level, message, data)
}

// 以 info 级别去打印所有的日志
func NewCompatibleLogger(l *L) *CompatibleLogger {
	return &CompatibleLogger{l}
}
