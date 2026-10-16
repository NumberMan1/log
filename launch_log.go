package serverlog

import (
	"log"
	"os"
	"strconv"
	"sync"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

var bootInfoPrinter func(v ...any)
var bootInfoMu sync.RWMutex

func init() {
	printer := log.New(os.Stderr, "BootLog: ", log.Ldate|log.Ltime).Println
	bootInfoPrinter = func(v ...any) {
		stack := stackTraceFrame(3)
		stackStr := ""
		if stack != nil {
			stackStr = " " + stack.File + ":" + strconv.Itoa(stack.Line) + ":"
		}
		v = append([]any{stackStr, ansiRed}, v...)
		v = append(v, ansiReset)
		printer(v...)
	}
}

// SetBootInfoPrinter 并发安全地设置启动日志函数；nil 表示禁用。
// 自定义函数可能被并发调用，应自行同步其状态。
func SetBootInfoPrinter(f func(v ...any)) {
	bootInfoMu.Lock()
	defer bootInfoMu.Unlock()
	bootInfoPrinter = f
}

// GetBootInfoPrinter 返回当前启动日志函数的快照，可能为 nil。
func GetBootInfoPrinter() func(v ...any) {
	bootInfoMu.RLock()
	defer bootInfoMu.RUnlock()
	return bootInfoPrinter
}

// BootInfo 使用独立启动输出函数记录消息，不要求默认 core 已初始化。
func BootInfo(v ...any) {
	if printer := GetBootInfoPrinter(); printer != nil {
		printer(v...)
	}
}
