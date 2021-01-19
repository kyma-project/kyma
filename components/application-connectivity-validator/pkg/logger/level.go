package logger

import (
	"errors"
	"fmt"

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

var all_levels = []Level{DEBUG, INFO, WARN, ERROR, FATAL}

func MapLevel(level string) (Level, error) {
	var lvl = Level(level)

	switch lvl {
	case DEBUG, INFO, WARN, ERROR, FATAL:
		return lvl, nil
	default:
		return lvl, errors.New(fmt.Sprintf("Given log level: %s, doesn't match with any of %v", level, all_levels))
	}
}

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
