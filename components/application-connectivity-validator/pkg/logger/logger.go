package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(format Format, level Level, additionalCores ...zapcore.Core) *Logger {
	defaultCore := zapcore.NewCore(
		format.toZapEncoder(),
		zapcore.Lock(os.Stderr),
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level <= level
		}),
	)
	cores := append(additionalCores, defaultCore)
	return &Logger{zap.New(zapcore.NewTee(cores...)).Sugar()}
}

func (l *Logger) WithFields(m map[string]string) *Logger {
	for key, val := range m {
		l.SugaredLogger = l.With(key, val)
	}
	return l
}

func (l *Logger) WithContext(context map[string]string) *Logger {
	l.SugaredLogger = l.With("context", context)
	return l
}

// By default the Fatal Error log will be in json format, because it's production default.
func LogFatalError(format string, args ...interface{}) {
	logger := New(JSON, ERROR)
	logger.Fatalf(format, args...)
}

// By default the Options log will be in json format, because it's production default.
func LogOptions(format string, args ...interface{}) {
	logger := New(JSON, INFO)
	logger.Infof(format, args...)
}
