package logging

import (
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
)

func ConfigureLogger(logLevel, logFormat string) (*logger.Logger, error) {
	parsedLogLevel, err := logger.MapLevel(logLevel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse logging level")
	}

	format, err := logger.MapFormat(logFormat)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set logging format")
	}

	l, err := logger.New(format, parsedLogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set logger")
	}

	if err := logger.InitKlog(l, parsedLogLevel); err != nil {
		return nil, errors.Wrap(err, "unable to init Klog")
	}

	return l, nil
}
