package step

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"strings"
)

type Retried struct {
	steps []Step
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
	retryOptions := []retry.Option{
		retry.Attempts(20), // at max (100 * (1 << 13)) / 1000 = 819,2 sec
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf(".")
		}),
	}
	return retry.Do(func() error {
		for _, step := range r.steps {
			err := step.Run()
			if err != nil {
				return err
			}
		}
		return nil
	}, retryOptions...)
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
