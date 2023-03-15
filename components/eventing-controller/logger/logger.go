package logger

import (
	"context"

	"github.com/kyma-project/kyma/common/logging/logger"
	"go.uber.org/zap"
)

type Logger struct {
	*logger.Logger
}

// New returns a new Kyma standardized Logger with the given format and level.
// This method needs to return Logger in favor of KLogger until the whole codebase is migrated to the new interface.
func New(format, level string) (*Logger, error) {
	log, err := build(format, level)
	return log.(*Logger), err
}

// KLogger is an interface for the Kyma wide logger.
type KLogger interface {
	WithContext() *zap.SugaredLogger
	WithTracing(ctx context.Context) *zap.SugaredLogger
}

func build(format, level string) (KLogger, error) {
	logFormat, err := logger.MapFormat(format)
	if err != nil {
		return nil, err
	}

	logLevel, err := logger.MapLevel(level)
	if err != nil {
		return nil, err
	}

	log, err := logger.New(logFormat, logLevel)
	if err != nil {
		return nil, err
	}

	if err = logger.InitKlog(log, logLevel); err != nil {
		return nil, err
	}

	return &Logger{Logger: log}, nil
}

func NewWithAtomicLevel(format, level string) (*Logger, error) {
	logFormat, err := logger.MapFormat(format)
	if err != nil {
		return nil, err
	}

	logLevel, err := logger.MapLevel(level)
	if err != nil {
		return nil, err
	}

	atomicLevel := zap.NewAtomicLevel()
	log, err := logger.NewWithAtomicLevel(logFormat, atomicLevel)
	if err != nil {
		return nil, err
	}

	if err = logger.InitKlog(log, logLevel); err != nil {
		return nil, err
	}

	return &Logger{Logger: log}, nil
}
