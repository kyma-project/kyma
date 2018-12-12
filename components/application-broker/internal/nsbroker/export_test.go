package nsbroker

import (
	"time"

	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
)

func (f *Facade) WithServiceChecker(serviceChecker serviceChecker) *Facade {
	f.serviceChecker = serviceChecker
	return f
}

func NewHTTPChecker(afterFunc func(d time.Duration) <-chan time.Time) *httpChecker {
	return &httpChecker{
		log: spy.NewLogDummy().Logger,
		afterFunc: afterFunc,
		sleepTime: time.Millisecond,
	}
}
