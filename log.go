package serverlog

import (
	"os"
	"sync/atomic"
)

var (
	gLogger      *L = NewLogger()
	gEventLogger *L = NewLogger()
)

// 为 fatal 等级提供可替换的退出函数，默认使用 os.Exit。
// 通过 atomic.Value 来保证并发读写安全，避免测试覆盖时产生数据竞争。
var exitFn atomic.Value

func init() {
	exitFn.Store(func(code int) { os.Exit(code) })
}

// SetExitFunc 允许在测试中覆盖退出行为，避免直接退出测试进程。
// 在生产环境中不建议调用该方法。
func SetExitFunc(fn func(int)) {
	if fn == nil {
		// 保护性兜底，避免设置为 nil 导致 Load 失败
		exitFn.Store(func(code int) { os.Exit(code) })
		return
	}
	exitFn.Store(fn)
}

// ResetExitFunc 重置退出行为为默认的 os.Exit。
func ResetExitFunc() {
	exitFn.Store(func(code int) { os.Exit(code) })
}

func NewLogger() *L {
	return &L{
		levelPrinters: make(map[Printer]*LevelSwitch),
		maxLevel:      TraceLevel,
	}
}

// // Init Debug和Info默认输出到标准输出；Event以JSON格式记录日志文件，按小时自动切换；Error和Fatal上报到Sentry后台
// func Init(dir, dsn, environment string) {
// 	InitSentry(dsn, environment)
// 	// 事件日志
// 	ep, err := NewEventPrinter(path.Join(dir, "event"))
// 	if err != nil {
// 		panic(err)
// 	}
// 	AddPrinter(NewLevelSwitch().EnableSingleLevel(EventLevel), ep)

// 	// 调试日志
// 	AddPrinter(NewLevelSwitch().EnableLevels(DebugLevel, InfoLevel), NewConsolePrinter())

// 	// 错误日志
// 	jp, err := NewJsonlogPrinter(path.Join(dir, "log"))
// 	if err != nil {
// 		panic(err)
// 	}
// 	AddPrinter(NewLevelSwitch().EnableLevelLte(ErrorLevel), jp)
// 	AddPrinter(NewLevelSwitch().EnableLevelLte(ErrorLevel), NewSentryPrinter())
// }

func SetDefaultLogger(l *L) {
	gLogger = l
}

func GetDefaultLogger() *L {
	return gLogger
}

func SetDefaultEventLogger(l *L) {
	gEventLogger = l
}

func GetDefaultEventLogger() *L {
	return gEventLogger
}

func AddEventPrinter(printer Printer) {
	gEventLogger.AddPrinter(NewLevelSwitch().EnableLevels(AllLevels...), printer)
}

func AddPrinter(levelSwitch *LevelSwitch, printer Printer) {
	gLogger.AddPrinter(levelSwitch, printer)
}

func RemovePrinter(printer Printer) {
	gLogger.RemovePrinter(printer)
}

func GetLevelSwitcherByPrinter(printer Printer) *LevelSwitch {
	return gLogger.GetLevelSwitcherByPrinter(printer)
}

// Close 关闭全局日志系统
func Close() {
	gLogger.Close()
	gEventLogger.Close()
}

func SetMaxLevel(level Level) {
	gLogger.SetMaxLevel(level)
}

func I(msg string) {
	gLogger.Info(msg, nil)
}

func D(msg string) {
	gLogger.Debug(msg, nil)
}

func E(msg string) {
	gLogger.Error(msg, nil)
}

func P(msg string) {
	gLogger.Panic(msg, nil)
}

// Fatal 上报致命错误到Sentry后台
func F(msg string) {
	gLogger.Fatal(msg, nil)
}

// Debug 输出调试信息到终端标准输出
func Debug(msg string, data M) {
	gLogger.Debug(msg, data)
}

// Info 输出信息到终端标准输出
func Info(msg string, data M) {
	gLogger.Info(msg, data)
}

// Event 输出事件信息到文件，用于定位bug
func Event(data M) {
	// 默认的事件打印器
	gEventLogger.Info("", data)
}

// Error 上报错误信息到Sentry后台
func Error(msg string, data M) {
	gLogger.Error(msg, data)
}

// Panic 上报严重错误到Sentry后台
func Panic(msg string, data M) {
	gLogger.Panic(msg, data)
}

// Fatal 上报致命错误到Sentry后台
func Fatal(msg string, data M) {
	gLogger.Fatal(msg, data)
}

// L 日志记录器
type L struct {
	levelPrinters map[Printer]*LevelSwitch
	maxLevel      Level
	preHandler    func(level Level, msg string, data M) (Level, string, M)
}

func (logger *L) Close() {
	for p := range logger.levelPrinters {
		p.Close()
	}
}

func (logger *L) SetPreHandler(preHandler func(level Level, msg string, data M) (Level, string, M)) {
	logger.preHandler = preHandler
}

func (logger *L) AddPrinter(levelSwitch *LevelSwitch, printer Printer) {
	logger.levelPrinters[printer] = levelSwitch
}

func (logger *L) RemovePrinter(printer Printer) {
	delete(logger.levelPrinters, printer)
}

func (logger *L) GetLevelSwitcherByPrinter(printer Printer) *LevelSwitch {
	return logger.levelPrinters[printer]
}

// M 数据Map别名
type M = map[string]interface{}

// Debug 输出调试信息到终端标准输出
func (logger *L) Debug(msg string, data M) {
	logger.Log(DebugLevel, msg, data)
}

// Info 输出信息到终端标准输出
func (logger *L) Info(msg string, data M) {
	logger.Log(InfoLevel, msg, data)
}

// Error 上报错误信息到Sentry后台
func (logger *L) Error(msg string, data M) {
	logger.Log(ErrorLevel, msg, data)
}

// Panic 上报严重错误到Sentry后台
func (logger *L) Panic(msg string, data M) {
	logger.Log(PanicLevel, msg, data)
}

// Fatal 上报致命错误到Sentry后台
func (logger *L) Fatal(msg string, data M) {
	logger.Log(FatalLevel, msg, data)
}

func (logger *L) SetMaxLevel(level Level) {
	logger.maxLevel = level
}

func (logger *L) Log(level Level, msg string, data M) {
	if level > logger.maxLevel {
		return
	}
	data = ErrorMsgPreHandler(level, msg, data)
	if logger.preHandler != nil {
		level, msg, data = logger.preHandler(level, msg, data)
	}
	for p, ls := range logger.levelPrinters {
		if ls.IsEnabled(level) {
			p.Log(level, msg, data)
		}
	}

	switch level {
	case PanicLevel:
		panic(msg)
	case FatalLevel:
		// 关闭自己
		logger.Close()
		// 关闭全局
		gLogger.Close()
		gEventLogger.Close()
		// 使用可替换的退出函数，默认 关闭全局logger 并 os.Exit(1)
		fn := exitFn.Load().(func(int))
		fn(1)
	}
}
