package serverlog

// OnlyMessageFormat 只输出消息，适合已有外部格式的 io.Writer 适配。
type OnlyMessageFormat struct {
}

// NewOnlyMessageFormat 创建可并发使用的纯消息格式器。
func NewOnlyMessageFormat() *OnlyMessageFormat {
	return &OnlyMessageFormat{}
}

// Format 原样返回 msg，忽略等级和 data。
func (obj *OnlyMessageFormat) Format(level Level, msg string, data M) (result string) {
	return msg
}
