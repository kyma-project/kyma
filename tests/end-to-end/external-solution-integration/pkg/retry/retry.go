package retry

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"
)

const retryAttemptsCount = 120
const retryDelay = 1 * time.Second

var defaultOpts = []retry.Option{
	retry.Attempts(retryAttemptsCount),
	retry.Delay(retryDelay),
	retry.DelayType(retry.FixedDelay),
	retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "RetryTest").Debugf("OnRetry: attempts: %d, error: %v", n, err)
	}),
}

func Do(fn retry.RetryableFunc, opts ...retry.Option) error {
	allOpts := append(defaultOpts, opts...)
	return retry.Do(fn, allOpts...)
}
