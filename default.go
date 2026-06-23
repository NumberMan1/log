package serverlog

import (
	"os"
	"strings"
)

func init() {
	var maxLevel Level
	if str := os.Getenv("MAX_LOG_LEVEL"); str != "" {
		var err error
		maxLevel, err = ParseLevel(str)
		if err != nil {
			panic(err)
		}
	} else {
		maxLevel = defaultMaxLevel
	}
	SetMaxLevel(maxLevel)
}

var inited bool

func Init(logPath string, sentryDsn string) {
	if inited {
		I("serverlog already inited, ignore")
		return
	}
	inited = true

	allConsole := os.Getenv("ALL_LOG_CONSOLE")
	// 非本地的时候应该发送到不同的地方
	consoleLevelSwitch := NewLevelSwitch()
	if strings.ToLower(allConsole) == "true" {
		consoleLevelSwitch.EnableLevels(AllLevels...)
	} else {
		consoleLevelSwitch.EnableLevels(InfoLevel, DebugLevel)
	}
	AddPrinter(
		consoleLevelSwitch,
		NewConsolePrinter(ConsolePrinterConfig{}),
	)

	if logPath != "" {
		filePrinter, err := NewFilePrinter(FilePrinterConfig{
			Dir:        logPath + "/log/",
			PreHandler: StacksWrapPreHandler,
		})
		if err != nil {
			panic(err)
		}
		AddPrinter(NewLevelSwitch().EnableLevels(ErrorLevel, FatalLevel), filePrinter)
	}

	if sentryDsn != "" {
		AddPrinter(LevelSwitchLte(ErrorLevel), NewSentryPrinter(SentryPrinterConfig{
			Dsn: sentryDsn,
		}))
	}

	if logPath != "" {
		// 事件日志记入文件
		eventPrinter, err := NewFilePrinter(FilePrinterConfig{
			Dir:        logPath + "/event/",
			PreHandler: AddTimePreHandler,
		})
		if err != nil {
			panic(err)
		}
		AddEventPrinter(eventPrinter)
	}
}
