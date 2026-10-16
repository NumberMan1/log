package serverlog

// ConcatPreHandler 按传入顺序组合处理函数，所有函数必须非 nil，前一个返回值作为后一个输入。
// 空列表保持原数据；组合函数的并发安全性取决于各处理函数。
func ConcatPreHandler(handlers ...PreHandler) func(level Level, msg string, data M) M {
	return func(level Level, msg string, data M) M {
		for _, handler := range handlers {
			data = handler(level, msg, data)
		}
		return data
	}
}
