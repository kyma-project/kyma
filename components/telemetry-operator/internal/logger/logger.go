package logger

import (
	"fmt"

	"github.com/kyma-project/kyma/common/logging/logger"
	"k8s.io/klog"
)

type Logger struct {
	*logger.Logger
}

// New returns a new logger with the given format and level.
func New(format, level string) (*Logger, error) {
	fmt.Println(format)
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

	klog.CopyStandardLogTo("ERROR")

	return &Logger{Logger: log}, nil
}
