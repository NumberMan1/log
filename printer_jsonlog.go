package serverlog

import (
	"github.com/NumberMan1/log/fileutil"
)

// NewJsonlogPrinter 创建按小时轮转的 JSON 输出器，记录消息、等级、时间和调用位置。
// 队列容量为 4000，初始化失败时返回错误。
func NewJsonlogPrinter(dir string) (*FilePrinter, error) {
	l, err := NewFilePrinter(FilePrinterConfig{
		Dir:         dir,
		Switcher:    fileutil.HOURS_SWITCHER,
		FileType:    ".log",
		LogChanSize: 4000,
		PreHandler:  StackWrapPreHandler,
	})
	if err != nil {
		return nil, err
	}

	return l, nil
}
