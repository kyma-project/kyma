package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(format Format, level Level, additionalCores ...zapcore.Core) *Logger {
	defaultCore := zapcore.NewCore(
		format.toZapEncoder(),
		zapcore.Lock(os.Stderr),
		level.toZapLevelEnabler(),
	)
	cores := append(additionalCores, defaultCore)
	return &Logger{zap.New(zapcore.NewTee(cores...)).Sugar()}
}

func (l *Logger) WithTracing(context map[string]string) *Logger {
	if val, ok := context["traceid"]; ok {
		l.SugaredLogger = l.With("traceid", val)
	} else {
		l.SugaredLogger = l.With("traceid", "unknown")
	}
	if val, ok := context["spanid"]; ok {
		l.SugaredLogger = l.With("spanid", val)
	} else {
		l.SugaredLogger = l.With("spanid", "unknown")
	}
	return l
}

func (l *Logger) WithContext(context map[string]string) *Logger {
	l.SugaredLogger = l.With("context", context)
	return l
}
