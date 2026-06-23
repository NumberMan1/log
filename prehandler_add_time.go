package serverlog

import "time"

// AddTimePreHandler 如果data为空就直接return，否则添加 Time
func AddTimePreHandler(level Level, msg string, data M) M {
	if data == nil {
		return data
	}
	data["Time"] = time.Now().Unix()
	return data
}
