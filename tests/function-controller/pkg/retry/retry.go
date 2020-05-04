package retry

import (
	goerrors "errors"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	DefaultBackoff = retry.DefaultBackoff
	ErrInvalidFunc = goerrors.New("invalid function")
)

func fnWithIgnore(fn func() error, ignoreErr func(error) bool, log shared.Logger) func() error {
	return func() error {
		if fn == nil {
			return ErrInvalidFunc
		}
		err := fn()
		if ignoreErr == nil {
			return err
		}
		if ignoreErr(err) {
			log.Logf("ignoring: %s", err)
			return nil
		}
		return err
	}
}

func errorFn(log shared.Logger) func(error) bool {
	return func(err error) bool {
		if errors.IsTimeout(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) {
			log.Logf("retrying due to: %s", err)
			return true
		}
		return false
	}
}

func WithIgnoreOnNotFound(backoff wait.Backoff, fn func() error, log shared.Logger) error {
	return retry.OnError(backoff, errorFn(log), fnWithIgnore(fn, errors.IsNotFound, log))
}

func WithIgnoreOnConflict(backoff wait.Backoff, fn func() error, log shared.Logger) error {
	return retry.OnError(backoff, errorFn(log), fnWithIgnore(fn, errors.IsConflict, log))
}

func OnTimeout(backoff wait.Backoff, fn func() error, log shared.Logger) error {
	return retry.OnError(backoff, errorFn(log), fn)
}
