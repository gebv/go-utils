package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Configure configure zap logger.
func Configure(levelSet string) {
	level := zapcore.InfoLevel
	if err := level.Set(levelSet); err != nil {
		panic(err)
	}
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(level)
	l, err := config.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
	zap.RedirectStdLog(l.Named("stdlog"))
}
