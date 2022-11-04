package logger

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/go-logr/zapr"
	"github.com/kyma-project/kyma/common/logging/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
)

type Logger struct {
	zapLogger *zap.SugaredLogger
}

/*
This function creates logger structure based on given format, atomicLevel and additional cores
AtomicLevel structure allows to change level dynamically
*/
func NewWithAtomicLevel(format Format, atomicLevel zap.AtomicLevel, additionalCores ...zapcore.Core) (*Logger, error) {
	return new(format, atomicLevel, additionalCores...)
}

/*
This function creates logger structure based on given format, level and additional cores
*/
func New(format Format, level Level, additionalCores ...zapcore.Core) (*Logger, error) {
	filterLevel, err := level.ToZapLevel()
	if err != nil {
		return nil, errors.Wrap(err, "while getting zap log level")
	}

	levelEnabler := zap.LevelEnablerFunc(func(incomingLevel zapcore.Level) bool {
		return incomingLevel >= filterLevel
	})

	return new(format, levelEnabler, additionalCores...)
}

func new(format Format, levelEnabler zapcore.LevelEnabler, additionalCores ...zapcore.Core) (*Logger, error) {
	encoder, err := format.ToZapEncoder()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting encoding configuration  for %s format", format)
	}

	defaultCore := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stderr),
		levelEnabler,
	)
	cores := append(additionalCores, defaultCore)
	return &Logger{zap.New(zapcore.NewTee(cores...), zap.AddCaller()).Sugar()}, nil
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

/*
By default the Fatal Error log will be in json format, because it's production default.
*/
func LogFatalError(format string, args ...interface{}) error {
	logger, err := New(JSON, ERROR)
	if err != nil {
		return errors.Wrap(err, "while getting Error Json Logger")
	}
	logger.zapLogger.Fatalf(format, args...)
	return nil
}

/*
This function initialize klog which is used in k8s/go-client
*/
func InitKlog(log *Logger, level Level) error {
	zaprLogger := zapr.NewLogger(log.WithContext().Desugar())
	lvl, err := level.ToZapLevel()
	if err != nil {
		return errors.Wrap(err, "while getting zap log level")
	}
	zaprLogger.V((int)(lvl))
	klog.SetLogger(zaprLogger)
	return nil
}
