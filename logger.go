package log

import (
	"github.com/NumberMan1/general/log"
	"github.com/NumberMan1/log/field"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

type Logger struct {
	base *zap.Logger
	fs   []field.Field
}

func (l *Logger) With(fields ...field.Field) log.Logger {
	return &Logger{
		base: l.base,
		fs:   append(l.fs, fields...),
	}
}

func (l *Logger) fields(f ...field.Field) []field.Field {
	if len(l.fs) != 0 {
		return append(f, l.fs...)
	}
	return f
}

func (l *Logger) Debug(msg string, field ...field.Field) {
	l.base.Debug(msg, l.fields(field...)...)
}

func (l *Logger) Info(msg string, field ...field.Field) {
	l.base.Info(msg, l.fields(field...)...)
}

func (l *Logger) Warn(msg string, field ...field.Field) {
	l.base.Warn(msg, l.fields(field...)...)
}

func (l *Logger) Error(msg string, field ...field.Field) {
	l.base.Error(msg, l.fields(field...)...)
}

func (l *Logger) Sync() {
	_ = l.base.Sync()
}

type builder struct {
	conf *config
}

func initLogger(opts ...Option) {
	logger = newLogger(opts...)
}

func newLogger(opts ...Option) log.Logger {
	bd := &builder{
		conf: newDefaultConfig(),
	}

	for i := range opts {
		opts[i](bd.conf)
	}

	lg := bd.build()
	if bd.conf.appName != "" {
		lg = lg.With(field.String("app", bd.conf.appName))
	}
	lg.(*Logger).Sync()
	lg.Debug("log init completed")
	return lg
}

func (l *builder) build() log.Logger {
	lg, err := l.conf.Build(l.cores())
	checkErr(err)
	return &Logger{base: lg.WithOptions(zap.AddCallerSkip(1))}
}

func (l *builder) cores() zap.Option {
	cores := make([]zapcore.Core, 0)
	cores = append(cores, l.fileoutCores()...)
	cores = append(cores, l.stdoutCores()...)

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

func (l *builder) fileoutCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0)
	if l.conf.toFile {
		fileEncoder := zapcore.NewJSONEncoder(l.conf.EncoderConfig)
		cores = append(cores,
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "error", l.conf.fileAsync), l.priority(zapcore.ErrorLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "warn", l.conf.fileAsync), l.priority(zapcore.WarnLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "info", l.conf.fileAsync), l.priority(zapcore.InfoLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "debug", l.conf.fileAsync), l.priority(zapcore.DebugLevel)),
		)
	}
	return cores
}

func (l *builder) stdoutCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0)
	if l.conf.stdout {
		stdoutEncoder := zapcore.NewConsoleEncoder(l.conf.EncoderConfig)
		if strings.ToLower(l.conf.stdoutType) == "json" {
			stdoutEncoder = zapcore.NewJSONEncoder(l.conf.EncoderConfig)
		}
		debugConsoleWS := zapcore.Lock(os.Stdout) // 控制台标准输出
		errorConsoleWS := zapcore.Lock(os.Stderr)
		cores = append(cores,
			zapcore.NewCore(stdoutEncoder, errorConsoleWS, l.priority(zapcore.ErrorLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.WarnLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.InfoLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.DebugLevel)),
		)
	}
	return cores
}

func (l *builder) priority(level zapcore.Level) zap.LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		return lvl == level && level >= l.conf.Level.Level()
	}
}
