package serverlog

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NumberMan1/log/fileutil"
)

type FilePrinterConfig struct {
	Dir             string
	Switcher        fileutil.Switcher // 文件切换选择器，目前支持 HOURS_SWITCHER 和 DAY_SWITCHER, 默认: HOURS_SWITCHER
	FileType        string            // 默认: .log
	WriteBufferSize int               // 默认: 4096
	FlushTick       time.Duration     // 默认: 2s
	LogChanSize     int               // 默认: 2000
	LogFormat       LogFormat         // 默认: json
	PreHandler      PreHandler        // 默认只返回 data，抛弃掉了level和msg，这相当于打印的时候是把 data 序列化为json输出
}

type FilePrinter struct {
	config     FilePrinterConfig
	logChan    chan string
	closeWait  sync.WaitGroup
	closeMutex sync.RWMutex
	closeFlag  int32
	file       *fileutil.File
	logFormat  LogFormat
	preHandler PreHandler
}

func NewFilePrinter(config FilePrinterConfig) (*FilePrinter, error) {
	if config.FileType == "" {
		config.FileType = ".log"
	}
	if config.FileType[0] != '.' {
		config.FileType = "." + config.FileType
	}

	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = 4096
	}

	if config.FlushTick <= 0 {
		config.FlushTick = 2 * time.Second
	}

	if config.LogChanSize <= 0 {
		config.LogChanSize = 2000
	}

	if config.LogFormat == nil {
		config.LogFormat = NewJsonFormat()
	}

	if config.Switcher == nil {
		config.Switcher = fileutil.HOURS_SWITCHER
	}

	if config.PreHandler == nil {
		config.PreHandler = func(level Level, msg string, data M) M { return data }
	}

	printer := &FilePrinter{
		config:     config,
		logChan:    make(chan string, config.LogChanSize),
		logFormat:  config.LogFormat,
		preHandler: config.PreHandler,
	}

	if err := printer.switchFile(); err != nil {
		return nil, err
	}

	printer.closeWait.Add(1)
	go printer.loop()

	return printer, nil
}

func (printer *FilePrinter) loop() {
	defer func() {
		if printer.file != nil {
			printer.file.Close()
		}
		printer.closeWait.Done()
	}()

	// 定时刷新文件
	flushTicker := time.NewTicker(printer.config.FlushTick)
	defer flushTicker.Stop()

	// 定时切换文件
	switchTimer := time.NewTimer(printer.config.Switcher.FirstSwitchTime())
	defer switchTimer.Stop()

	for {
		select {
		case r, ok := <-printer.logChan:
			if ok {
				printer.file.Write(r)
				printer.file.WriteBytes(line)
			} else {
				for r := range printer.logChan {
					printer.file.Write(r)
					printer.file.WriteBytes(line)
				}
				return
			}
		case <-flushTicker.C:
			if err := printer.file.Flush(); err != nil {
				log.Println("file flush failed:", err.Error())
				panic(err)
			}
		case <-switchTimer.C:
			if err := printer.switchFile(); err != nil {
				log.Println("file switch failed:", err.Error())
				panic(err)
			}
			switchTimer.Reset(printer.config.Switcher.NextSwitchTime())
		}
	}
}

// 切换文件
func (printer *FilePrinter) switchFile() error {
	dir, fileName := printer.config.Switcher.DirAndFileName(printer.config.Dir)

	fullName := fileName + ".01" + printer.config.FileType
	for fileID := 2; fileutil.FileExists(fullName); fileID++ {
		fullName = fileName + fmt.Sprintf(".%02d", fileID) + printer.config.FileType
	}

	// 确认目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 先关闭旧文件再切换
	if printer.file != nil {
		if err := printer.file.Close(); err != nil {
			return err
		}
	}

	// 创建或者打开已存在文件
	file, err := fileutil.NewFile(fullName, printer.config.WriteBufferSize)
	if err != nil {
		return err
	}
	printer.file = file
	return nil
}

// 关闭日志系统
func (printer *FilePrinter) Close() {
	if atomic.LoadInt32(&printer.closeFlag) == 1 {
		return
	}

	printer.closeMutex.Lock()
	defer printer.closeMutex.Unlock()

	if atomic.LoadInt32(&printer.closeFlag) == 1 {
		return
	}

	atomic.StoreInt32(&printer.closeFlag, 1)
	close(printer.logChan)
	printer.closeWait.Wait()

	printer.file.Close()
}

// 在日志文件中输出信息
func (printer *FilePrinter) Log(level Level, message string, data M) {
	if atomic.LoadInt32(&printer.closeFlag) == 1 {
		return
	}

	printer.closeMutex.RLock()
	defer printer.closeMutex.RUnlock()

	if atomic.LoadInt32(&printer.closeFlag) == 1 {
		return
	}

	if printer.preHandler != nil {
		data = printer.preHandler(level, message, data)
	}

	printer.logChan <- printer.logFormat.Format(level, message, data)
}
