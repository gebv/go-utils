package logger

import (
	"os"

	zapsentry "github.com/gebv/go-utils/zap-sentry"
	"github.com/getsentry/raven-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc/grpclog"
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

type LoggerConfig struct {
	SentryDSN               string
	EnableDevelopmentLogger bool
	EnableDebugLevelLogger  bool
	Version                 string
}

func SetLogger(s *LoggerConfig) func() error {
	dsn := s.SentryDSN
	host, err := os.Hostname()
	if err != nil {
		zap.L().Panic("Failed to get hostname.", zap.Error(err))
	}

	if err := raven.SetDSN(dsn); err != nil {
		zap.L().Panic("Failed to set Sentry DSN.", zap.Error(err))
	}

	raven.CaptureMessageAndWait("Start application", nil)

	sentryCore, err := zapsentry.Configuration{
		DSN: dsn,
		Tags: map[string]string{
			"host": host,
		},
		Release: s.Version,
	}.Build()
	if err != nil {
		zap.L().Panic("Failed to create Sentry logger.", zap.Error(err))
	}

	// stop using early startup logger
	zap.L().Sync()

	// create logger config (production or development)
	var config zap.Config
	if s.EnableDevelopmentLogger {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig() // replace with zap.NewProductionConfig() when we are ready to use it
		config.Development = false
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	if s.EnableDebugLevelLogger {
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config.Level.SetLevel(zap.InfoLevel)
	}

	// create logger from config and Sentry wrapper
	l, err := config.Build(
		zap.AddStacktrace(zap.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, sentryCore)
		}),
	)
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(l)
	zap.RedirectStdLog(l.Named("stdlog"))

	grpclog.SetLogger(zapgrpc.NewLogger(l, zapgrpc.WithDebug()))

	if config.Development {
		zap.L().Info("Development logger initiated.")
	} else {
		zap.L().Info("Production logger initiated.")
	}

	return l.Sync
}
