package configurelogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel struct {
	Atomic  zap.AtomicLevel
	Default string
}

func New(atomic zap.AtomicLevel) *LogLevel {
	var l LogLevel
	l.Atomic = atomic
	l.Default = atomic.String()
	return &l
}

func (l *LogLevel) SetDefaultLogLevel() error {
	return l.ChangeLogLevel(l.Default)
}

func (l *LogLevel) ChangeLogLevel(logLevel string) error {
	parsedLevel, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	l.Atomic.SetLevel(parsedLevel)
	return nil
}
