package logger

import (
	"github.com/kyma-project/kyma/common/logging/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
)

type Logger struct {
	*logger.Logger
}

// New returns a new logger with the given format and level.
func New(format, level string, atomic zap.AtomicLevel) (*Logger, error) {
	logFormat, err := logger.MapFormat(format)
	if err != nil {
		return nil, err
	}

	logLevel, err := logger.MapLevel(level)
	if err != nil {
		return nil, err
	}

	log, err := logger.NewWithAtomicLevel(logFormat, atomic)
	if err != nil {
		return nil, err
	}

	if err = logger.InitKlog(log, logLevel); err != nil {
		return nil, err
	}

	// Redirects logs those are being written using standard logging mechanism to klog
	// to avoid logs from controller-runtime being pushed to the standard logs.
	klog.CopyStandardLogTo("ERROR")

	return &Logger{Logger: log}, nil
}

type LogLevel struct {
	Atomic  zap.AtomicLevel
	Default string
}

func NewLogReconfigurer(atomic zap.AtomicLevel) *LogLevel {
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
