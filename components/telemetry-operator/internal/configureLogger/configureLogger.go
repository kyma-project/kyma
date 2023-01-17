package configureLogger

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

func (l *LogLevel) ReconfigureLogLevel(config map[string]interface{}) error {
	if logLevel, ok := config["logLevel"].(string); ok {
		return l.ChangeLogLevel(logLevel)
	}
	// We would need to set the log level back when we remove the override config
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
