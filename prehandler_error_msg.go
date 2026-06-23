package serverlog

import (
	"fmt"
)

var errorKey = []string{"error", "err"}

// ErrorMsgPreHandler 调用 error.Error() 得到字符串, FormatJson 和 FormatText 硬编码支持
func ErrorMsgPreHandler(level Level, msg string, data M) M {
	if data == nil {
		return data
	}
	for _, key := range errorKey {
		if e, ok := data[key]; ok {
			data[key] = extraErrorMsg(e)
		}
	}
	return data
}

func extraErrorMsg(e any) string {
	if e == nil {
		return ""
	}
	if err, ok := e.(error); ok {
		return err.Error()
	}

	return fmt.Sprintf("%v", e)
}
