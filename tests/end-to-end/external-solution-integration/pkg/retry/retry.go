package retry

import (
	"time"

	"github.com/avast/retry-go"
)

const retryAttemptsCount = 120
const retryDelay = 1 * time.Second

var defaultOpts = []retry.Option{
	retry.Attempts(retryAttemptsCount),
	retry.Delay(retryDelay),
	retry.DelayType(retry.FixedDelay),
}

func Do(fn retry.RetryableFunc) error {
	return retry.Do(fn, defaultOpts...)
}

func WithCustomOpts(fn retry.RetryableFunc, opts ...retry.Option) error {
	allOpts := append(defaultOpts, opts...)

	return retry.Do(fn, allOpts...)
}
