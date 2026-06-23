package serverlog

func ConcatPreHandler(handlers ...PreHandler) func(level Level, msg string, data M) M {
	return func(level Level, msg string, data M) M {
		for _, handler := range handlers {
			data = handler(level, msg, data)
		}
		return data
	}
}
