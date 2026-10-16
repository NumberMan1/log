package serverlog

import "sync"

var initMu sync.Mutex
var inited bool

// Init is the legacy one-time initializer. Use Configure for errors/reconfiguration.
func Init(logPath, sentryDsn string) {
	initMu.Lock()
	defer initMu.Unlock()
	if inited {
		return
	}
	if err := Configure(Config{Stdout: true, LogFilePath: logPath, SentryDSN: sentryDsn}); err != nil {
		panic(err)
	}
	inited = true
}

func init() {
	level, err := configuredLevel("")
	if err != nil {
		panic(err)
	}
	SetMaxLevel(level)
}
