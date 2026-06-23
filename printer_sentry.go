package serverlog

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	jsoniter "github.com/json-iterator/go"
)

type ExtraTagFunc func(tags map[string]string)

type SentryPrinterConfig struct {
	Dsn          string
	Environment  string
	Release      string
	ExtraTagFunc ExtraTagFunc     // 默认为空，主要用于比如GameServer可以添加tag来标识日志归属于哪个sid
	PreHandler   PreHandler       // 默认为去除掉日志等级和msg，sentry将只能查到 server.M 部分
	Transport    sentry.Transport // 测试可注入 Transport，避免对 sentry.CaptureEvent 进行打补丁
	FlushTimeout time.Duration    // 默认为 2s，用于退出前尽量发送事件
}

type SentryPrinter struct {
	client       *sentry.Client
	extraTagFunc ExtraTagFunc
	preHandler   PreHandler
	flushTimeout time.Duration
	closeOnce    sync.Once
}

func NewSentryPrinter(config SentryPrinterConfig) *SentryPrinter {
	environment := config.Environment
	if environment == "" {
		environment = os.Getenv("SENTRY_ENVIRONMENT")
	}
	if environment == "" {
		environment = os.Getenv("SERVER_TYPE")
	}
	if environment == "" {
		environment = "develop"
	}

	release := config.Release
	if release == "" {
		release = os.Getenv("SENTRY_RELEASE")
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              config.Dsn,
		AttachStacktrace: true,
		Environment:      environment,
		Release:          release,
		ServerName:       os.Getenv("SERVER_NAME"),
		Transport:        config.Transport,
	})
	if err != nil {
		panic(err)
	}

	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 2 * time.Second
	}

	result := &SentryPrinter{
		client:       client,
		flushTimeout: config.FlushTimeout,
	}

	if config.ExtraTagFunc == nil {
		result.extraTagFunc = func(_ map[string]string) {}
	} else {
		result.extraTagFunc = config.ExtraTagFunc
	}

	if config.PreHandler == nil {
		result.preHandler = func(_ Level, _ string, m M) M { return m }
	} else {
		result.preHandler = config.PreHandler
	}

	return result
}

func (l *SentryPrinter) Log(level Level, msg string, data M) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Sentry error:%v msg:%s data:%v \n", err, msg, data)
		}
	}()

	event := sentry.NewEvent()

	l.extraTagFunc(event.Tags)

	data = l.preHandler(level, msg, data)

	// event.Tags["sid"] = gUserID
	event.Level = sentry.Level(level.String())
	event.Message = msg
	event.Extra = map[string]interface{}{
		"time": time.Now().Unix(),
	}
	if data != nil {
		event.Extra["Data"], _ = jsoniter.MarshalToString(data)
		if pid, ok := data["pid"]; ok {
			pidStr := fmt.Sprint(pid)
			event.User.ID = pidStr
		}
		for k, v := range data {
			event.Tags[k] = fmt.Sprint(v)
		}
	}

	event.Threads = []sentry.Thread{{
		Stacktrace: sentry.NewStacktrace(),
		Crashed:    false,
		Current:    true,
	}}

	l.client.CaptureEvent(event, nil, sentry.CurrentHub().Scope())
}

func (l *SentryPrinter) Close() {
	l.closeOnce.Do(func() {
		if l.client == nil {
			return
		}
		_ = l.client.Flush(l.flushTimeout)
	})
}
