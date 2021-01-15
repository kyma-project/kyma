package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TRACE_KEY = "traceid"
	SPAN_KEY  = "spanid"
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

func (l *Logger) WithTracing(ctx context.Context) *Logger {

	return l.enhanceLogger(ctx, TRACE_KEY, "unknown").
		enhanceLogger(ctx, SPAN_KEY, "unknown")
}

func (l *Logger) enhanceLogger(ctx context.Context, key, defaultValue string) *Logger {
	newLogger := &Logger{}
	val := ctx.Value(key)

	if val, ok := val.(string); ok {
		newLogger.SugaredLogger = l.With(key, val)
	} else {
		newLogger.SugaredLogger = l.With(key, "unknown")
	}

	return newLogger
}

func (l *Logger) WithContext(context map[string]string) *Logger {
	l.SugaredLogger = l.With("context", context)
	return l
}

/*
By default the error log will be in json format, because it's production default.
*/
func LogFatalError(format string, args ...interface{}) {
	logger := New(JSON, ERROR)
	logger.Fatalf(format, args...)
}
