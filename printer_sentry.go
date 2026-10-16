package serverlog

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"sync"
	"time"
)

// ExtraTagFunc 追加或覆盖输出事件的 tags，可能并发调用。
type ExtraTagFunc func(tags map[string]string)

// SentryPrinterConfig 配置独立 Sentry Client、字段映射和关闭行为。
type SentryPrinterConfig struct {
	// Dsn 非空时覆盖 ClientOptions.Dsn；环境变量回退由 SDK 管理。
	Dsn string
	// Environment 非空时覆盖 ClientOptions.Environment，否则读取 SENTRY_ENVIRONMENT。
	Environment string
	// Release 非空时覆盖 ClientOptions.Release，否则读取 SENTRY_RELEASE。
	Release string
	// ExtraTagFunc 在字段映射完成后追加或覆盖 tag；nil 表示不额外修改。
	ExtraTagFunc ExtraTagFunc
	// PreHandler 在映射前处理字段的顶层副本；nil 表示直接使用原始字段。
	PreHandler PreHandler
	// Transport 非 nil 时覆盖 ClientOptions.Transport，可用于离线测试或自定义发送。
	Transport sentry.Transport
	// ClientOptions 传递其余 SDK 选项，如 ServerName、BeforeSend 和重试行为。
	ClientOptions sentry.ClientOptions
	// FlushTimeout 是关闭时等待发送的上限；非正值采用 2 秒。
	FlushTimeout time.Duration
	// FieldMapper 非 nil 时完全替代默认映射规则，实现在并发调用中必须安全。
	FieldMapper SentryFieldMapper
	// TagKeys 显式指定转换为 tags 的字段；默认不映射任何字段。
	TagKeys []string
	// UserIDKeys 按顺序选择首个非空、非零值作为 User.ID；默认不推导用户身份。
	UserIDKeys []string
	// ExceptionKeys 指定原始 error 字段；nil 默认为 error、err，空切片禁用。
	ExceptionKeys []string
	// ExcludeKeys 指定在映射前移除的字段；默认不移除任何字段。
	ExcludeKeys []string
	// MaxFieldLength 是附加字段字符串的最大字节数；非正值采用 4096，保持有效 UTF-8。
	MaxFieldLength int
	// OnError 同步报告映射/提交 panic 和 Flush 超时，nil 时写标准日志。
	// 不得重入当前 Printer 或关闭其所属 core；回调 panic 会被捕获。
	OnError func(error)
}

// SentryPrinter 使用独立 Client/Scope 上报日志，Log 与 Close 可并发调用。
type SentryPrinter struct {
	client       *sentry.Client
	mapper       SentryFieldMapper
	extraTagFunc ExtraTagFunc
	preHandler   PreHandler
	flushTimeout time.Duration
	onError      func(error)
	mu           sync.RWMutex
	closed       bool
	closeOnce    sync.Once
	errMu        sync.RWMutex
	lastError    error
}

// NewSentryPrinter 创建输出器，非法配置时 panic。
// 需要错误返回时使用 NewSentryPrinterWithError。
func NewSentryPrinter(config SentryPrinterConfig) *SentryPrinter {
	printer, err := NewSentryPrinterWithError(config)
	if err != nil {
		panic(err)
	}
	return printer
}

// NewSentryPrinterWithError 创建独立 Client，配置错误以 error 返回。
// 仅采用 Sentry SDK 标准环境变量，不读取应用自定义环境变量。
func NewSentryPrinterWithError(config SentryPrinterConfig) (*SentryPrinter, error) {
	options := config.ClientOptions
	if config.Dsn != "" {
		options.Dsn = config.Dsn
	}
	if config.Environment != "" {
		options.Environment = config.Environment
	}
	if options.Environment == "" {
		options.Environment = os.Getenv("SENTRY_ENVIRONMENT")
	}
	if config.Release != "" {
		options.Release = config.Release
	}
	if options.Release == "" {
		options.Release = os.Getenv("SENTRY_RELEASE")
	}
	if config.Transport != nil {
		options.Transport = config.Transport
	}
	options.AttachStacktrace = true
	client, err := sentry.NewClient(options)
	if err != nil {
		return nil, err
	}
	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 2 * time.Second
	}
	mapper := config.FieldMapper
	if mapper == nil {
		mapper = NewDefaultSentryFieldMapper(config)
	}
	return &SentryPrinter{
		client: client, mapper: mapper, extraTagFunc: config.ExtraTagFunc,
		preHandler: config.PreHandler, flushTimeout: config.FlushTimeout, onError: config.OnError,
	}, nil
}

// Log 同步构造并提交事件，网络传输由 SDK 管理；mapper/SDK panic 会转为可观察错误。
// 关闭后的调用被忽略，原始输入不会被顶层字段处理修改。
func (printer *SentryPrinter) Log(level Level, msg string, data M) {
	// A mapper/SDK failure must not terminate the application.
	defer func() {
		if value := recover(); value != nil {
			printer.reportError(fmt.Errorf("sentry log: %v", value))
		}
	}()
	printer.mu.RLock()
	defer printer.mu.RUnlock()
	if printer.closed {
		return
	}
	data = cloneData(data)
	if printer.preHandler != nil {
		data = printer.preHandler(level, msg, data)
	}
	now := time.Now()
	fields := printer.mapper.Map(SentryLogEvent{Level: level, Message: msg, Data: data, Time: now})
	event := sentry.NewEvent()
	event.Timestamp = now
	event.Message = msg
	event.Level = sentryLevel(level)
	event.User = fields.User
	event.Exception = fields.Exception
	for key, value := range fields.Tags {
		event.Tags[key] = value
	}
	if printer.extraTagFunc != nil {
		printer.extraTagFunc(event.Tags)
	}
	for key, value := range fields.Contexts {
		event.Contexts[key] = value
	}
	// sentry-go v0.47 removed Event.Extra; keep payload structured in Contexts.
	if len(fields.Extra) != 0 {
		event.Contexts["extra"] = sentry.Context(fields.Extra)
	}
	if len(event.Exception) == 0 {
		event.Threads = []sentry.Thread{{Stacktrace: sentry.NewStacktrace(), Current: true}}
	}
	// Keep each logger independent of the global SDK hub's user and tags.
	printer.client.CaptureEvent(event, nil, sentry.NewScope())
}

func sentryLevel(level Level) sentry.Level {
	switch level {
	case PanicLevel, FatalLevel:
		return sentry.LevelFatal
	case ErrorLevel:
		return sentry.LevelError
	case WarnLevel:
		return sentry.LevelWarning
	case InfoLevel:
		return sentry.LevelInfo
	default:
		return sentry.LevelDebug
	}
}

// Err 返回最近一次 mapper/提交 panic 或关闭 Flush 超时，尚未发生时为 nil。
// SDK 后台网络重试状态由 Transport 管理，不包含在此返回值中。
func (printer *SentryPrinter) Err() error {
	printer.errMu.RLock()
	defer printer.errMu.RUnlock()
	return printer.lastError
}
func (printer *SentryPrinter) reportError(err error) {
	printer.errMu.Lock()
	printer.lastError = err
	printer.errMu.Unlock()
	reportPrinterError(printer.onError, err)
}

// Close 等待在途提交完成，限时 Flush 后释放 Client/Transport；重复调用安全。
// 不可在本输出器的 mapper、预处理或错误回调中重入。
func (printer *SentryPrinter) Close() {
	printer.closeOnce.Do(func() {
		printer.mu.Lock()
		printer.closed = true
		printer.mu.Unlock()
		defer printer.client.Close()
		if !printer.client.Flush(printer.flushTimeout) {
			printer.reportError(fmt.Errorf("sentry flush timed out after %s", printer.flushTimeout))
		}
	})
}

func reportPrinterError(handler func(error), err error) {
	defer func() {
		if value := recover(); value != nil {
			log.Printf("log error handler panic: %v", value)
		}
	}()
	if handler != nil {
		handler(err)
	} else {
		log.Printf("log output error: %v", err)
	}
}
