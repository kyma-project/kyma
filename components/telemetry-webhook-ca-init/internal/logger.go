package internal

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/kyma-project/kyma/common/logging/logger"
)

func InitLogger() logr.Logger {
	logFormat, err := logger.MapFormat("text")
	if err != nil {
		panic(err)
	}
	logLevel, err := logger.MapLevel("info")
	if err != nil {
		panic(err)
	}
	kymaLogger, err := logger.New(logFormat, logLevel)
	if err != nil {
		panic(err)
	}
	if err = logger.InitKlog(kymaLogger, logLevel); err != nil {
		panic(err)
	}

	return zapr.NewLogger(kymaLogger.WithContext().Desugar())
}
