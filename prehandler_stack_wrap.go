package serverlog

import "time"

// StackWrapPreHandler 包装消息、等级、Unix 秒时间和首个库外调用位置，原 data 放入 Data。
func StackWrapPreHandler(level Level, message string, data M) M {
	m := map[string]interface{}{
		"Time":    time.Now().Unix(),
		"LogType": level.String(),
		"Message": message,
		"Stack":   applicationFrame(),
	}
	if data != nil {
		m["Data"] = data
	}
	return m
}

// StacksWrapPreHandler 包装消息、等级、Unix 秒时间和库外调用栈，原 data 放入 Data。
func StacksWrapPreHandler(level Level, message string, data M) M {
	m := map[string]interface{}{
		"Time":    time.Now().Unix(),
		"LogType": level.String(),
		"Message": message,
		"Stack":   applicationFrames(),
	}
	if data != nil {
		m["Data"] = data
	}
	return m
}

// NewStackWrapPreHandler 创建按 skipFrame 解析栈的包装函数，供自定义调用层校准。
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
