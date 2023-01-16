package ConfigureLogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel struct {
	Atomic zap.AtomicLevel
}

func New(atomic zap.AtomicLevel) *LogLevel {
	var l LogLevel
	l.Atomic = atomic
	return &l
}

func (l *LogLevel) ChangeLogLevel(logLevel string) error {
	parsedLevel, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	l.Atomic.SetLevel(parsedLevel)
	return nil
}
