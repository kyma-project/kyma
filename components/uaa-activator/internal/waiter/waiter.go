package waiter

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/ctxutil"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Default constraints
const (
	Interval = time.Second
	Timeout  = 5 * time.Minute
)

// ConditionFunc returns nil error if the condition is satisfied, or an error
// if should be repeated.
type ConditionFunc func() (err error)

// WaitForSuccess waits for function until it returns nil error or timeout occurs
func WaitForSuccess(ctx context.Context, condition ConditionFunc) error {
	var lastErr error
	err := wait.PollImmediate(Interval, Timeout, func() (done bool, err error) {
		if ctxutil.ShouldExit(ctx) {
			return false, ctx.Err()
		}

		if err := condition(); err != nil {
			lastErr = err
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return errors.Wrapf(err, "got error while waiting for condition, last error: %v", lastErr)
	}

	return nil
}
