package serverlog

import (
	"github.com/NumberMan1/log/fileutil"
)

func NewEventPrinter(dir string) (*FilePrinter, error) {
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
