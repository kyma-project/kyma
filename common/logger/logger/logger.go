package logger

import (
	"context"
	"github.com/kyma-project/kyma/common/logger/tracing"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
)

type Logger struct {
	zapLogger *zap.SugaredLogger
}

func New(format Format, level Level, additionalCores ...zapcore.Core) *Logger {
	filterLevel := level.toZapLevel()

	defaultCore := zapcore.NewCore(
		format.ToZapEncoder(),
		zapcore.Lock(os.Stderr),
		zap.LevelEnablerFunc(func(incomingLevel zapcore.Level) bool {
			return incomingLevel >= filterLevel
		}),
	)
	cores := append(additionalCores, defaultCore)
	return &Logger{zap.New(zapcore.NewTee(cores...)).Sugar()}
}

func (l *Logger) WithTracing(ctx context.Context) *zap.SugaredLogger {
	newLogger := *l
	for key, val := range tracing.GetMetadata(ctx) {
		newLogger.zapLogger = newLogger.zapLogger.With(key, val)
	}

	return newLogger.WithContext()
}

func (l *Logger) WithContext() *zap.SugaredLogger {
	return l.zapLogger.With(zap.Namespace("context"))
}

/**
By default the Fatal Error log will be in json format, because it's production default.
*/
func LogFatalError(format string, args ...interface{}) {
	logger := New(JSON, ERROR)
	logger.zapLogger.Fatalf(format, args...)
}

/**
This function initialize klog which is used in k8s/go-client
*/
func InitKlog(log *Logger, level Level) {
	zaprLogger := zapr.NewLogger(log.WithContext().Desugar())
	zaprLogger.V((int)(level.toZapLevel()))
	klog.SetLogger(zaprLogger)

}
