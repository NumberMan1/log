package serverlog

import "os"

type Builder struct {
	l *L
}

func NewBuilder() *Builder {
	return &Builder{
		l: NewLogger(),
	}
}

// 起到跟原来的Jsonlog一样的打印效果: 每个小时一个文件，日志内容会添加栈后包裹一层，适用于日志打印
func (b *Builder) AddJsonlogPrinter(maxLevel Level, dir string) *Builder {
	printer, err := NewJsonlogPrinter(dir)
	if err != nil {
		panic(err)
	}
	return b.AddCustomPrinter(maxLevel, printer)
}

// 将日志以json格式打印到文件, 适用于数据上报
//
// config 如果不做任何配置，打印效果将是: 每个小时一个文件，日志内容是将 data 序列化为json字符串输出
func (b *Builder) AddFilePrinter(maxLevel Level, config FilePrinterConfig) *Builder {
	printer, err := NewFilePrinter(config)
	if err != nil {
		panic(err)
	}
	return b.AddCustomPrinter(maxLevel, printer)
}

// 将日志以类似unix日志格式打印到终端，适用于调试时日志查看
func (b *Builder) AddConsolePrinter(maxLevel Level, config ConsolePrinterConfig) *Builder {
	printer := NewConsolePrinter(config)
	return b.AddCustomPrinter(maxLevel, printer)
}

// dsn: sentry 的API地址, environment: sentry环境， 这两个参数主要是用于初始化Sentry时
func (b *Builder) AddSentryPrinter(maxLevel Level, config SentryPrinterConfig) *Builder {
	if config.Environment == "" {
		config.Environment = os.Getenv("SERVER_TYPE")
	}
	printer := NewSentryPrinter(config)
	return b.AddCustomPrinter(maxLevel, printer)
}

func (b *Builder) AddCustomPrinter(maxLevel Level, printer Printer) *Builder {
	b.l.AddPrinter(LevelSwitchLte(maxLevel), printer)
	return b
}

func (b *Builder) Logger() *L {
	return b.l
}
