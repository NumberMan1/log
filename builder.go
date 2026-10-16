package serverlog

// Builder 在启动阶段逐步构建一个独立 core，自身不支持并发修改。
// 需要可恢复的初始化错误时使用 NewConfiguredLogger 或各输出器的 error 构造函数。
type Builder struct {
	l *L
}

// NewBuilder 创建没有输出器的构建器。
func NewBuilder() *Builder {
	return &Builder{
		l: NewLogger(),
	}
}

// AddJsonlogPrinter 注册按小时轮转、包含调用位置的 JSON 文件输出；失败时 panic。
func (b *Builder) AddJsonlogPrinter(maxLevel Level, dir string) *Builder {
	printer, err := NewJsonlogPrinter(dir)
	if err != nil {
		panic(err)
	}
	return b.AddCustomPrinter(maxLevel, printer)
}

// AddFilePrinter 注册可配置文件输出；默认只序列化 data，初始化失败时 panic。
func (b *Builder) AddFilePrinter(maxLevel Level, config FilePrinterConfig) *Builder {
	printer, err := NewFilePrinter(config)
	if err != nil {
		panic(err)
	}
	return b.AddCustomPrinter(maxLevel, printer)
}

// AddConsolePrinter 注册终端输出器，并启用 maxLevel 及更高严重程度的等级。
func (b *Builder) AddConsolePrinter(maxLevel Level, config ConsolePrinterConfig) *Builder {
	printer := NewConsolePrinter(config)
	return b.AddCustomPrinter(maxLevel, printer)
}

// AddSentryPrinter 注册独立 Sentry 输出器；不推导应用环境，非法配置时 panic。
func (b *Builder) AddSentryPrinter(maxLevel Level, config SentryPrinterConfig) *Builder {
	printer := NewSentryPrinter(config)
	return b.AddCustomPrinter(maxLevel, printer)
}

// AddCustomPrinter 注册调用方提供的输出器，其关闭责任转交给构建的 core。
func (b *Builder) AddCustomPrinter(maxLevel Level, printer Printer) *Builder {
	b.l.AddPrinter(LevelSwitchLte(maxLevel), printer)
	return b
}

// Logger 返回构建中的同一个 core，不复制配置或转移到默认 logger。
func (b *Builder) Logger() *L {
	return b.l
}
