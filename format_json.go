package serverlog

import (
	"encoding/json"
	"fmt"
)

type JsonFormat struct {
}

func NewJsonFormat() *JsonFormat {
	return &JsonFormat{}
}

func (obj *JsonFormat) Format(level Level, msg string, data M) (result string) {
	if data == nil {
		return ""
	}

	var seqBytes []byte
	var err error
	seqBytes, err = json.Marshal(data)
	if err != nil {
		seqBytes = []byte(fmt.Sprintf(`{"error":"%s", "msg":"%s", "data":"%#v"}`, err.Error(), msg, data))
	}
	return string(seqBytes)
}
