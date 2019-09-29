package step

import (
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
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
	return retry.Do(func() error {
		for _, step := range r.steps {
			err := step.Run()
			if err != nil {
				return err
			}
		}
		return nil
	})
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
