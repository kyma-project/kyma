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

var allLevels = []Level{DEBUG, INFO, WARN, ERROR, FATAL}

func MapLevel(level string) (Level, error) {
	var lvl = Level(level)

	switch lvl {
	case DEBUG, INFO, WARN, ERROR, FATAL:
		return lvl, nil
	default:
		return lvl, errors.New(fmt.Sprintf("Given log level: %s, doesn't match with any of %v", level, allLevels))
	}
}

func (l Level) ToZapLevel() (zapcore.Level, error) {
	switch l {
	case DEBUG:
		return zap.DebugLevel, nil
	case INFO:
		return zap.InfoLevel, nil
	case WARN:
		return zap.WarnLevel, nil
	case ERROR:
		return zap.ErrorLevel, nil
	case FATAL:
		return zap.FatalLevel, nil
	default:
		return zap.DebugLevel, errors.New("unknown level")
	}
}
