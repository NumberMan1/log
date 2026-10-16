package serverlog

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TextFormat 输出等级、时间、调用位置、消息以及按键排序的字段。
type TextFormat struct {
	skipFrame int
}

// NewTextFormat 创建文本格式器；skipFrame < 0 时自动查找库外调用位置。
// 非负值直接传给运行时栈解析，供自定义包装层校准。
func NewTextFormat(skipFrame int) *TextFormat {
	return &TextFormat{
		skipFrame: skipFrame,
	}
}

// Format 将日志格式化为单行文本，不修改传入字段。
func (obj *TextFormat) Format(level Level, msg string, data M) (result string) {
	f := stackTraceFrame(obj.skipFrame)
	if obj.skipFrame < 0 {
		f = applicationFrame()
	}
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
