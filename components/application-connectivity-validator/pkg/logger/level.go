package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level string

const (
	DEBUG Level = "debug"
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
	FATAL Level = "fatal"
)

func (l Level) toZapLevel() zapcore.Level {
	switch l {
	case DEBUG:
		return zap.DebugLevel
	case INFO:
		return zap.InfoLevel
	case WARN:
		return zap.WarnLevel
	case ERROR:
		return zap.ErrorLevel
	case FATAL:
		return zap.FatalLevel
	default:
		panic("unknown level")
	}
}

func (l Level) toZapLevelEnabler() zap.LevelEnablerFunc {
	return func(zl zapcore.Level) bool {
		return zl <= l.toZapLevel()
	}
}
