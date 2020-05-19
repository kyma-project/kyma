package retry

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"
)

const retryAttemptsCount = 120
const retryDelay = 1 * time.Second

type Option = retry.Option

var Attempts = retry.Attempts
var Delay = retry.Delay
var DelayType = retry.DelayType
var OnRetry = retry.OnRetry
var FixedDelay = retry.FixedDelay
var BackOffDelay = retry.BackOffDelay

var defaultOpts = []retry.Option{
	Attempts(retryAttemptsCount),
	Delay(retryDelay),
	DelayType(FixedDelay),
	OnRetry(func(n uint, err error) {
		logrus.WithField("component", "RetryTest").Debugf("OnRetry: attempts: %d, error: %v", n, err)
	}),
}

func Do(fn retry.RetryableFunc, opts ...retry.Option) error {
	allOpts := append(defaultOpts, opts...)
	return retry.Do(fn, allOpts...)
}
