package serverlog

import (
	"os"
	"sync"
	"sync/atomic"
)

var (
	defaultMu sync.RWMutex
	gLogger   *L = NewLogger()
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

// NewLogger 创建不带输出器的独立 core，默认允许 Trace 及更高严重程度的日志。
// 调用方负责注册 Printer，并在生命周期结束时调用 Close。
func NewLogger() *L {
	return &L{
		levelPrinters: make(map[Printer]*LevelSwitch),
		maxLevel:      TraceLevel,
	}
}

// SetDefaultLogger 并发安全地替换默认 core，不关闭旧 core；传入 nil 会创建空 core。
// 需要自动关闭旧资源时使用 Configure。
func SetDefaultLogger(l *L) {
	if l == nil {
		l = NewLogger()
	}
	defaultMu.Lock()
	gLogger = l
	defaultMu.Unlock()
}

// GetDefaultLogger 返回当前默认 core 的快照；后续替换不会改变已取得的引用。
func GetDefaultLogger() *L {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return gLogger
}

// AddPrinter 为默认 core 注册输出器，具体生命周期规则见 L.AddPrinter。
func AddPrinter(levelSwitch *LevelSwitch, printer Printer) {
	GetDefaultLogger().AddPrinter(levelSwitch, printer)
}

// RemovePrinter 从默认 core 解除注册，不关闭输出器，也不取消已开始的打印。
func RemovePrinter(printer Printer) {
	GetDefaultLogger().RemovePrinter(printer)
}

// GetLevelSwitcherByPrinter 查询已注册的等级开关；未注册时返回 nil。
func GetLevelSwitcherByPrinter(printer Printer) *LevelSwitch {
	return GetDefaultLogger().GetLevelSwitcherByPrinter(printer)
}

// Close 等待当前 core 的在途日志分发并关闭输出器；并发或重复调用安全。
// 不可在该 core 的 PreHandler、Printer.Log 或错误回调中重入 Close。
func Close() {
	GetDefaultLogger().Close()
}

// SetMaxLevel 并发安全地设置允许输出的最低严重程度；数值较大的等级被过滤。
func SetMaxLevel(level Level) {
	GetDefaultLogger().SetMaxLevel(level)
}

// I 使用默认 core 记录不带字段的 Info 日志。
func I(msg string) {
	GetDefaultLogger().Info(msg, nil)
}

// D 使用默认 core 记录不带字段的 Debug 日志。
func D(msg string) {
	GetDefaultLogger().Debug(msg, nil)
}

// E 使用默认 core 记录不带字段的 Error 日志。
func E(msg string) {
	GetDefaultLogger().Error(msg, nil)
}

// P 使用默认 core 记录日志后触发 panic。
func P(msg string) {
	GetDefaultLogger().Panic(msg, nil)
}

// F 使用默认 core 记录日志、关闭输出器，然后以状态码 1 退出。
func F(msg string) {
	GetDefaultLogger().Fatal(msg, nil)
}

// Debug 记录调试日志；实际去向由注册的 Printer 决定。
func Debug(msg string, data M) {
	GetDefaultLogger().Debug(msg, data)
}

// Info 记录普通信息；实际去向由注册的 Printer 决定。
func Info(msg string, data M) {
	GetDefaultLogger().Info(msg, data)
}

// Error 记录错误信息，不自动退出或指定某个上报后端。
func Error(msg string, data M) {
	GetDefaultLogger().Error(msg, data)
}

// Panic 尝试输出后触发 panic，即使日志被过滤也保留该控制流。
func Panic(msg string, data M) {
	GetDefaultLogger().Panic(msg, data)
}

// Fatal 尝试输出并关闭当前及默认 core 后退出进程；等级过滤不会阻止退出。
func Fatal(msg string, data M) {
	GetDefaultLogger().Fatal(msg, data)
}

// L 是并发安全的等级路由 core，拥有已注册输出器的关闭责任。
// 创建后不要按值复制；派生的字段 logger 共享其生命周期。
type L struct {
	mu            sync.RWMutex
	closed        bool
	closeOnce     sync.Once
	active        sync.WaitGroup
	levelPrinters map[Printer]*LevelSwitch
	maxLevel      Level
	preHandler    func(level Level, msg string, data M) (Level, string, M)
}

// Close 等待当前 core 的在途日志分发并关闭输出器；并发或重复调用安全。
// 不可在该 core 的 PreHandler、Printer.Log 或错误回调中重入 Close。
func (logger *L) Close() {
	logger.closeOnce.Do(func() {
		logger.mu.Lock()
		logger.closed = true
		printers := make([]Printer, 0, len(logger.levelPrinters))
		for p := range logger.levelPrinters {
			printers = append(printers, p)
		}
		logger.mu.Unlock()
		// 先在同一配置锁下拒绝新调用，再等待计数归零，避免 Add 与 Wait 的零值竞争。
		logger.active.Wait()
		for _, p := range printers {
			p.Close()
		}
	})
}

// SetPreHandler 并发安全地替换分发前的处理函数，nil 表示禁用。
// 函数可能被并发调用，接收独立的顶层字段 map，不应修改共享嵌套对象或关闭当前 core。
func (logger *L) SetPreHandler(preHandler func(Level, string, M) (Level, string, M)) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.preHandler = preHandler
}

// AddPrinter 注册输出器并接管其关闭责任；替换同一实例时更新等级开关。
// nil 参数或已关闭的 core 会忽略本次注册，输出器应使用可比较的具体类型。
func (logger *L) AddPrinter(levelSwitch *LevelSwitch, printer Printer) {
	if printer == nil || levelSwitch == nil {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if logger.closed {
		return
	}
	if logger.levelPrinters == nil {
		logger.levelPrinters = make(map[Printer]*LevelSwitch)
	}
	logger.levelPrinters[printer] = levelSwitch
}

// RemovePrinter 从默认 core 解除注册，不关闭输出器，也不取消已开始的打印。
func (logger *L) RemovePrinter(printer Printer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	delete(logger.levelPrinters, printer)
}

// GetLevelSwitcherByPrinter 查询已注册的等级开关；未注册时返回 nil。
func (logger *L) GetLevelSwitcherByPrinter(printer Printer) *LevelSwitch {
	logger.mu.RLock()
	defer logger.mu.RUnlock()
	return logger.levelPrinters[printer]
}

// M 是结构化字段 map 的别名；嵌套值按只读 payload 传递。
type M = map[string]interface{}

// Debug 记录调试日志；实际去向由注册的 Printer 决定。
func (logger *L) Debug(msg string, data M) {
	logger.Log(DebugLevel, msg, data)
}

// Info 记录普通信息；实际去向由注册的 Printer 决定。
func (logger *L) Info(msg string, data M) {
	logger.Log(InfoLevel, msg, data)
}

// Error 记录错误信息，不自动退出或指定某个上报后端。
func (logger *L) Error(msg string, data M) {
	logger.Log(ErrorLevel, msg, data)
}

// Panic 尝试输出后触发 panic，即使日志被过滤也保留该控制流。
func (logger *L) Panic(msg string, data M) {
	logger.Log(PanicLevel, msg, data)
}

// Fatal 尝试输出并关闭当前及默认 core 后退出进程；等级过滤不会阻止退出。
func (logger *L) Fatal(msg string, data M) {
	logger.Log(FatalLevel, msg, data)
}

// SetMaxLevel 并发安全地设置允许输出的最低严重程度；数值较大的等级被过滤。
func (logger *L) SetMaxLevel(level Level) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.maxLevel = level
}

// Trace 记录比 Debug 更细粒度的诊断信息。
func (logger *L) Trace(msg string, data M) { logger.Log(TraceLevel, msg, data) }

// Warn 记录需要关注但尚未终止当前流程的信息。
func (logger *L) Warn(msg string, data M) { logger.Log(WarnLevel, msg, data) }

// Trace 记录比 Debug 更细粒度的诊断信息。
func Trace(msg string, data M) { GetDefaultLogger().Trace(msg, data) }

// Warn 记录需要关注但尚未终止当前流程的信息。
func Warn(msg string, data M) { GetDefaultLogger().Warn(msg, data) }

// Log 根据最大等级和输出器开关分发日志，不提前将 error 转为字符串。
// 每次调用及每个输出器各获独立顶层 map；自定义回调必须自行保证并发安全。
func (logger *L) Log(level Level, msg string, data M) {
	originalLevel := level
	logger.mu.RLock()
	enabled := !logger.closed && level <= logger.maxLevel
	handler := logger.preHandler
	type output struct {
		p  Printer
		ls *LevelSwitch
	}
	outputs := make([]output, 0, len(logger.levelPrinters))
	if enabled {
		logger.active.Add(1)
		for p, ls := range logger.levelPrinters {
			outputs = append(outputs, output{p, ls})
		}
	}
	// 用户回调在锁外运行，允许安全查询/调整配置；active 仍保护关闭顺序。
	logger.mu.RUnlock()
	if enabled {
		func() {
			defer logger.active.Done()
			data = cloneData(data)
			if handler != nil {
				level, msg, data = handler(level, msg, data)
			}
			for _, o := range outputs {
				if o.ls.IsEnabled(level) {
					o.p.Log(level, msg, cloneData(data))
				}
			}
		}()
	}
	// Fatal/Panic control flow must survive filtering or a prehandler level change.
	switch originalLevel {
	case PanicLevel:
		panic(msg)
	case FatalLevel:
		logger.Close()
		Close()
		exitFn.Load().(func(int))(1)
	}
}

// cloneData isolates top-level fields. Nested payloads must be treated as immutable.
func cloneData(data M) M {
	if data == nil {
		return nil
	}
	result := make(M, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}
