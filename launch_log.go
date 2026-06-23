package serverlog

import (
	"log"
	"os"
	"strconv"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

var bootInfoPrinter func(v ...any)

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

func SetBootInfoPrinter(f func(v ...any)) {
	bootInfoPrinter = f
}

func GetBootInfoPrinter() func(v ...any) {
	return bootInfoPrinter
}

func BootInfo(v ...any) {
	bootInfoPrinter(v...)
}
