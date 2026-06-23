package serverlog

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

type TextFormat struct {
	skipFrame int
}

func NewTextFormat(skipFrame int) *TextFormat {
	return &TextFormat{
		skipFrame: skipFrame,
	}
}

func (obj *TextFormat) Format(level Level, msg string, data M) (result string) {
	f := stackTraceFrame(obj.skipFrame)
	b := &bytes.Buffer{}
	b.WriteString(fmt.Sprintf("[%s]\t%s\t%s:%d\t%s", strings.ToUpper(level.String())[:4], time.Now().Format(time.RFC3339), f.File, f.Line, msg))
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteByte(' ')
		b.WriteString(key)
		b.WriteByte('=')

		stringVal, ok := data[key].(string)
		if !ok {
			stringVal = fmt.Sprint(data[key])
		}

		b.WriteString(fmt.Sprintf("%q", stringVal))
	}

	return b.String()
}
