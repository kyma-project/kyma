package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func New(format Format, level Level) *zap.SugaredLogger {
	core := zapcore.NewTee(zapcore.NewCore(
		format.toZapEncoder(),
		zapcore.Lock(os.Stderr),
		level.toZapLevelEnabler(),
	))

	return zap.New(core).Sugar()
}

type Tracing struct {
	trace_id string
	span_id  string
}

func AddTracing(logger *zap.SugaredLogger, tracing Tracing) *zap.SugaredLogger {
	return logger.With(tracing)
}
