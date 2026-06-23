package serverlog

type OnlyMessageFormat struct {
}

func NewOnlyMessageFormat() *OnlyMessageFormat {
	return &OnlyMessageFormat{}
}

func (obj *OnlyMessageFormat) Format(level Level, msg string, data M) (result string) {
	return msg
}
