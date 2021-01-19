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
	filterLevel := level.ToZapLevel()

	defaultCore := zapcore.NewCore(
		format.toZapEncoder(),
		zapcore.Lock(os.Stderr),
		zap.LevelEnablerFunc(func(incomingLevel zapcore.Level) bool {
			return incomingLevel >= filterLevel
		}),
	)
	cores := append(additionalCores, defaultCore)
	return &Logger{zap.New(zapcore.NewTee(cores...)).Sugar()}
}

func (l *Logger) WithFields(m map[string]string) *Logger {
	newLogger := *l
	for key, val := range m {
		newLogger.SugaredLogger = newLogger.With(key, val)
	}
	return &newLogger
}

func (l *Logger) WithContext(context map[string]string) *Logger {
	newLogger := *l
	newLogger.SugaredLogger = newLogger.With(zap.Namespace("context"))
	return newLogger.EnhanceContext(context)
}

func (l *Logger) EnhanceContext(context map[string]string) *Logger {
	newLogger := *l
	for key, val := range context {
		newLogger.SugaredLogger = l.With(key, val)
	}
	return &newLogger
}

// By default the Fatal Error log will be in json format, because it's production default.
func LogFatalError(format string, args ...interface{}) {
	logger := New(JSON, ERROR)
	logger.Fatalf(format, args...)
}

// By default the Options log will be in json format, because it's production default.
func LogOptions(format string, args ...interface{}) {
	logger := New(JSON, INFO)
	logger.Debugf(format, args...)
}
