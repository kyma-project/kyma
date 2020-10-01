package retry

import (
	"time"

	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
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
}

func Do(fn retry.RetryableFunc, opts ...retry.Option) error {
	allOpts := append(defaultOpts, opts...)
	allOpts = append(allOpts, OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }))
	allOpts = append(allOpts, retry.LastErrorOnly(false))
	return retry.Do(fn, allOpts...)
}
