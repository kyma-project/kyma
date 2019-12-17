package repeat

import (
	"context"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/ctxutil"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ConditionFunc returns nil error if the condition is satisfied, or an error
// if should be repeated.
type ConditionFunc func() (err error)

// UntilSuccess repeats given condition as long as it returns an error.
func UntilSuccess(ctx context.Context, condition ConditionFunc) error {
	var lastErr error
	err := wait.PollImmediate(config.Interval, config.Timeout, func() (done bool, err error) {
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
