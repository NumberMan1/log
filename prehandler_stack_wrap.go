package serverlog

import "time"

// StackWrapPreHandler 在data外面包一层数据，添加 Time、LogType、Stack 的打印，方便用于调试
func StackWrapPreHandler(level Level, message string, data M) M {
	m := map[string]interface{}{
		"Time":    time.Now().Unix(),
		"LogType": level.String(),
		"Message": message,
		"Stack":   stackTraceFrame(6),
	}
	if data != nil {
		m["Data"] = data
	}
	return m
}

// StackWrapPreHandler 在data外面包一层数据，添加 Time、LogType、Stack 的打印，方便用于调试
func StacksWrapPreHandler(level Level, message string, data M) M {
	m := map[string]interface{}{
		"Time":    time.Now().Unix(),
		"LogType": level.String(),
		"Message": message,
		"Stack":   stackTraceFrames(6),
	}
	if data != nil {
		m["Data"] = data
	}
	return m
}

func NewStackWrapPreHandler(skipFrame int) func(level Level, message string, data M) M {
	return func(level Level, message string, data M) M {
		m := map[string]interface{}{
			"Time":    time.Now().Unix(),
			"LogType": level.String(),
			"Message": message,
			"Stack":   stackTraceFrames(skipFrame),
		}
		if data != nil {
			m["Data"] = data
		}
		return m
	}
}
