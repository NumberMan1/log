package serverlog

import (
	"fmt"
)

var errorKey = []string{"error", "err"}

// ErrorMsgPreHandler 在副本中把 error/err 键转换为字符串，保留其他字段。
// 这是显式的输出适配器；core 不会提前调用，以保留原始错误链。
func ErrorMsgPreHandler(level Level, msg string, data M) M {
	if data == nil {
		return data
	}
	data = cloneData(data)
	for _, key := range errorKey {
		if e, ok := data[key]; ok {
			data[key] = extraErrorMsg(e)
		}
	}
	return data
}

func extraErrorMsg(e any) string {
	if isNilValue(e) {
		return ""
	}
	if err, ok := e.(error); ok {
		return err.Error()
	}

	return fmt.Sprintf("%v", e)
}
