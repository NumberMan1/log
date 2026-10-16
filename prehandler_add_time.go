package serverlog

import "time"

// AddTimePreHandler 为非 nil data 的副本增加 Unix 秒时间戳 Time；nil 输入仍返回 nil。
func AddTimePreHandler(level Level, msg string, data M) M {
	if data == nil {
		return data
	}
	data = cloneData(data)
	data["Time"] = time.Now().Unix()
	return data
}
