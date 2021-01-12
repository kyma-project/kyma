package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func New(format Format, level Level) *zap.SugaredLogger {
	core := zapcore.NewCore(
		format.toZapEncoder(),
		zapcore.Lock(os.Stderr),
		level.toZapLevelEnabler(),
	)
	return newWithCustomCores(format, level, core)
}

func newWithCustomCores(format Format, level Level, cores ...zapcore.Core) *zap.SugaredLogger {
	return zap.New(zapcore.NewTee(cores...)).Sugar()
}
