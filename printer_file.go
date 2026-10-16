package serverlog

import (
	"errors"
	"fmt"
	"github.com/NumberMan1/log/fileutil"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// OverflowPolicy 定义异步文件队列已满时调用方的行为。
type OverflowPolicy uint8

const (
	// BlockOnFull preserves queued entries and applies backpressure (default).
	BlockOnFull OverflowPolicy = iota
	// DropNewest keeps the caller responsive and reports ErrLogQueueFull.
	DropNewest
)

// ErrLogQueueFull 表示 DropNewest 策略丢弃了一条尚未入队的日志。
var ErrLogQueueFull = errors.New("log queue full")

// FilePrinterConfig 配置目录、轮转、缓冲、格式和故障报告；零值采用字段注释中的默认值。
type FilePrinterConfig struct {
	// Dir 是轮转文件的根目录；具体日期子目录由 Switcher 决定。
	Dir string
	// Switcher 为 nil 时按小时轮转。
	Switcher fileutil.Switcher
	// FileType 默认为 .log，缺少点号时自动补齐。
	FileType string
	// WriteBufferSize 是缓冲字节数，非正值采用 4096。
	WriteBufferSize int
	// FlushTick 是刷新间隔，非正值采用 2 秒。
	FlushTick time.Duration
	// LogChanSize 是队列条数，非正值采用 2000。
	LogChanSize int
	// LogFormat 为 nil 时采用 JsonFormat，只序列化 data。
	LogFormat LogFormat
	// PreHandler 在序列化前处理字段；nil 表示不处理，不隐式补充 msg 或 level。
	PreHandler PreHandler
	// OverflowPolicy 默认为 BlockOnFull；可用 DropNewest 换取调用方不阻塞。
	OverflowPolicy OverflowPolicy
	// OnError 同步报告队列和 I/O 故障，nil 时写标准日志。
	// 回调须快速返回，不可重入当前 Printer 或关闭其所属 core；回调 panic 会被捕获。
	OnError func(error)
}

// FilePrinter 将格式化后的日志放入有界队列，由单个后台协程写入和轮转。
// Close 会排空已接受的队列；持续存储故障通过 Err/OnError 报告，不保证故障期间的数据持久化。
type FilePrinter struct {
	config     FilePrinterConfig
	logChan    chan string
	done       chan struct{}
	closeMutex sync.RWMutex
	closed     bool
	file       *fileutil.File
	logFormat  LogFormat
	preHandler PreHandler
	errMutex   sync.RWMutex
	lastError  error
	dropped    atomic.Uint64
}

// NewFilePrinter 校验配置、独占创建首个文件，再启动写入协程。
// 初始化失败不会启动后台协程，成功后调用方必须 Close。
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
	if config.OverflowPolicy != BlockOnFull && config.OverflowPolicy != DropNewest {
		return nil, fmt.Errorf("invalid log overflow policy: %d", config.OverflowPolicy)
	}
	printer := &FilePrinter{
		config: config, logChan: make(chan string, config.LogChanSize), done: make(chan struct{}),
		logFormat: config.LogFormat, preHandler: config.PreHandler,
	}
	if err := printer.switchFile(); err != nil {
		return nil, err
	}
	go printer.loop()
	return printer, nil
}

func (printer *FilePrinter) loop() {
	defer close(printer.done)
	defer func() {
		if err := printer.file.Close(); err != nil {
			printer.reportError(err)
		}
	}()
	flushTicker := time.NewTicker(printer.config.FlushTick)
	defer flushTicker.Stop()
	switchTimer := time.NewTimer(printer.config.Switcher.FirstSwitchTime())
	defer switchTimer.Stop()
	for {
		select {
		case record, ok := <-printer.logChan:
			if !ok {
				return
			}
			if err := printer.file.Write(record + "\n"); err != nil {
				printer.reportError(fmt.Errorf("write log: %w", err))
			}
		case <-flushTicker.C:
			if err := printer.file.Flush(); err != nil {
				printer.reportError(fmt.Errorf("flush log: %w", err))
			}
		case <-switchTimer.C:
			if err := printer.switchFile(); err != nil {
				printer.reportError(fmt.Errorf("rotate log: %w", err))
			}
			switchTimer.Reset(printer.config.Switcher.NextSwitchTime())
		}
	}
}

func (printer *FilePrinter) switchFile() error {
	dir, name := printer.config.Switcher.DirAndFileName(printer.config.Dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var next *fileutil.File
	for index := 1; ; index++ {
		var err error
		next, err = fileutil.NewExclusiveFile(fmt.Sprintf("%s.%02d%s", name, index, printer.config.FileType), printer.config.WriteBufferSize)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return err
		}
		break
	}
	// Keep the old file alive until its replacement has opened successfully.
	old := printer.file
	printer.file = next
	if old != nil {
		return old.Close()
	}
	return nil
}

// Close 停止接受新日志，排空已入队数据并关闭文件。
// 所有并发调用都会等待同一次关闭完成；不得在自己的错误回调中调用。
func (printer *FilePrinter) Close() {
	printer.closeMutex.Lock()
	if !printer.closed {
		printer.closed = true
		close(printer.logChan)
	}
	printer.closeMutex.Unlock()
	<-printer.done
}

// Log 在调用方协程中执行格式化，再按 OverflowPolicy 入队。
// 默认满队列时阻塞；关闭后的调用忽略，自定义格式器和预处理必须并发安全。
func (printer *FilePrinter) Log(level Level, message string, data M) {
	printer.closeMutex.RLock()
	closed := printer.closed
	printer.closeMutex.RUnlock()
	if closed {
		return
	}
	data = cloneData(data)
	if printer.preHandler != nil {
		data = printer.preHandler(level, message, data)
	}
	record := printer.logFormat.Format(level, message, data)
	printer.closeMutex.RLock()
	if printer.closed {
		printer.closeMutex.RUnlock()
		return
	}
	dropped := false
	if printer.config.OverflowPolicy == DropNewest {
		select {
		case printer.logChan <- record:
		default:
			dropped = true
		}
	} else {
		printer.logChan <- record
	}
	printer.closeMutex.RUnlock()
	if dropped {
		printer.dropped.Add(1)
		printer.reportError(ErrLogQueueFull)
	}
}

// Err 返回最近一次队列、写入、刷新、轮转或关闭错误，尚未发生时为 nil。
func (printer *FilePrinter) Err() error {
	printer.errMutex.RLock()
	defer printer.errMutex.RUnlock()
	return printer.lastError
}

// Dropped 返回因 DropNewest 满队列而丢弃的累计条数，不包含存储故障丢失。
func (printer *FilePrinter) Dropped() uint64 { return printer.dropped.Load() }
func (printer *FilePrinter) reportError(err error) {
	printer.errMutex.Lock()
	printer.lastError = err
	printer.errMutex.Unlock()
	reportPrinterError(printer.config.OnError, err)
}
