package step

import (
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
)

type Retried struct {
	steps   []Step
	options []Option
}

func (r *Retried) Name() string {
	names := make([]string, len(r.steps))
	for i, step := range r.steps {
		names[i] = step.Name()
	}
	joined := strings.Join(names, ", ")
	return fmt.Sprintf("Retried: %s", joined)
}

func (r *Retried) Run() error {
	return retry.Do(func() error {
		for _, step := range r.steps {
			err := step.Run()
			if err != nil {
				return err
			}
		}
		return nil
	}, r.options...)
}

func (r *Retried) Cleanup() error {
	var errAll *multierror.Error
	for i := len(r.steps) - 1; i >= 0; i-- {
		err := r.steps[i].Cleanup()
		errAll = multierror.Append(errAll, err)
	}
	return errAll.ErrorOrNil()
}

func Retry(steps ...Step) *Retried {
	return &Retried{steps: steps}
}

func (r *Retried) WithRetryOptions(options ...Option) *Retried {
	r.options = options
	return r
}

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
	return retry.Do(fn, allOpts...)
}
