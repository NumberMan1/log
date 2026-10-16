package serverlog

import (
	"github.com/NumberMan1/log/fileutil"
)

// NewDataPrinter 创建按小时轮转的数据输出器，队列容量为 4000，给非 nil data 增加 Time。
// 它不创建额外全局 logger，也不推导数据的业务含义。
func NewDataPrinter(dir string) (*FilePrinter, error) {
	l, err := NewFilePrinter(FilePrinterConfig{
		Dir:         dir,
		Switcher:    fileutil.HOURS_SWITCHER,
		FileType:    ".log",
		LogChanSize: 4000,
		PreHandler:  AddTimePreHandler,
	})
	if err != nil {
		return nil, err
	}

	return l, nil
}
