package serverlog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config 只描述通用输出和公共字段，不推导应用身份或业务字段。
// 零值不安装任何输出器；Fields 中的嵌套对象必须保持只读。
type Config struct {
	// Level 优先于 MAX_LOG_LEVEL；两者都为空时使用构建标签的默认等级。
	Level string `json:"level" yaml:"level"`
	// Stdout 是否安装标准输出 Printer，默认关闭。
	Stdout bool `json:"stdout" yaml:"stdout"`
	// StdoutTyp 允许 text、json 或空串；空串采用 text。
	StdoutTyp string `json:"stdout_typ" yaml:"stdout-typ"`
	// LogFilePath 是直接写入的日志根目录；空串禁用文件输出，不创建隐式业务子目录。
	LogFilePath string `json:"log_file_path" yaml:"log-file-path"`
	// SentryDSN 非空时覆盖 Sentry 配置中的 DSN，并启用错误上报。
	SentryDSN string `json:"sentry_dsn" yaml:"sentry-dsn"`
	// SentryEnvironment 非空时覆盖高级配置的环境。
	SentryEnvironment string `json:"sentry_environment" yaml:"sentry-environment"`
	// SentryRelease 非空时覆盖高级配置的版本。
	SentryRelease string `json:"sentry_release" yaml:"sentry-release"`
	// Fields 是显式公共字段；构造时复制顶层 map，本次日志的字段可以覆盖它们。
	// msg、level、date、caller 由元数据处理器维护，其中显式 caller 会保留。
	Fields M `json:"fields,omitempty" yaml:"fields,omitempty"`
	// Sentry 提供 SDK、映射、Transport 和故障回调配置，不参与 JSON/YAML 编解码。
	Sentry SentryPrinterConfig `json:"-" yaml:"-"`
}

func configuredLevel(value string) (Level, error) {
	if value == "" {
		value = os.Getenv("MAX_LOG_LEVEL")
	}
	if value == "" {
		return defaultMaxLevel, nil
	}
	return ParseLevel(value)
}

// NewConfiguredLogger 根据配置创建独立 core；调用方负责 Close。
// 初始化失败时关闭已创建的输出器并返回错误，不修改默认 logger。
func NewConfiguredLogger(config Config) (*L, error) {
	level, err := configuredLevel(config.Level)
	if err != nil {
		return nil, err
	}
	if config.StdoutTyp != "" && config.StdoutTyp != "text" && config.StdoutTyp != "json" {
		return nil, fmt.Errorf("invalid stdout type: %q", config.StdoutTyp)
	}
	raw := NewLogger()
	raw.SetMaxLevel(level)
	raw.SetPreHandler(metadataHandler(config.Fields))
	if config.Stdout {
		console := ConsolePrinterConfig{}
		if config.StdoutTyp == "json" {
			console.LogFormat = NewJsonFormat()
		}
		raw.AddPrinter(LevelSwitchLte(level), NewConsolePrinter(console))
	}
	if config.LogFilePath != "" {
		printer, err := NewFilePrinter(FilePrinterConfig{Dir: config.LogFilePath, PreHandler: StacksWrapPreHandler})
		if err != nil {
			raw.Close()
			return nil, err
		}
		raw.AddPrinter(LevelSwitchLte(level), printer)
	}
	sentryConfig := config.Sentry
	if config.SentryDSN != "" {
		sentryConfig.Dsn = config.SentryDSN
	}
	if config.SentryEnvironment != "" {
		sentryConfig.Environment = config.SentryEnvironment
	}
	if config.SentryRelease != "" {
		sentryConfig.Release = config.SentryRelease
	}
	if sentryConfig.Dsn != "" || sentryConfig.ClientOptions.Dsn != "" {
		printer, err := NewSentryPrinterWithError(sentryConfig)
		if err != nil {
			raw.Close()
			return nil, err
		}
		raw.AddPrinter(LevelSwitchLte(ErrorLevel), printer)
	}
	return raw, nil
}

func applicationCaller() string {
	frame := applicationFrame()
	if frame.File == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(frame.File)), filepath.Base(frame.File), frame.Line)
}

// metadataHandler snapshots configured fields; per-call fields take precedence.
func metadataHandler(fields M) func(Level, string, M) (Level, string, M) {
	defaults := cloneData(fields)
	return func(level Level, msg string, data M) (Level, string, M) {
		result := cloneData(defaults)
		if result == nil {
			result = M{}
		}
		for key, value := range data {
			result[key] = value
		}
		result["msg"], result["level"], result["date"] = msg, level.String(), time.Now().Format(time.RFC3339Nano)
		if _, exists := result["caller"]; !exists {
			result["caller"] = applicationCaller()
		}
		return level, msg, result
	}
}

// Configure 在所有输出初始化成功后原子替换默认 core，再关闭旧 core。
// 失败时保留旧默认值；已有 scope 仍引用旧 core，重新配置后应重新获取。
func Configure(config Config) error {
	raw, err := NewConfiguredLogger(config)
	if err != nil {
		return err
	}
	defaultMu.Lock()
	old := gLogger
	gLogger = raw
	defaultMu.Unlock()
	old.Close()
	return nil
}
