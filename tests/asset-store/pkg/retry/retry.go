package retry

import (
	"fmt"

	goerrors "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	DefaultBackoff      = retry.DefaultBackoff
	ErrInvalidErrorFunc = goerrors.New("invalid error function")
)

func fnWithIgnore(fn func() error, ignoreErr func(error) bool, callbacks ...func(...interface{})) func() error {
	return func() error {
		if fn == nil {
			return ErrInvalidErrorFunc
		}
		err := fn()
		if ignoreErr == nil {
			return err
		}
		if ignoreErr(err) {
			for _, callback := range callbacks {
				msg := fmt.Sprintf("ignoring: %s", err)
				callback(msg)
			}
			return nil
		}
		return err
	}
}

func errorFn(callbacks ...func(...interface{})) func(error) bool {
	return func(err error) bool {
		if errors.IsTimeout(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) {
			for _, callback := range callbacks {
				msg := fmt.Sprintf("retrying due to: %s", err)
				callback(msg)
			}
			return true
		}
		return false
	}
}

func OnCreateError(backoff wait.Backoff, fn func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, errorFn(callbacks...), fnWithIgnore(fn, errors.IsAlreadyExists, callbacks...))
}

func OnDeleteError(backoff wait.Backoff, fn func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, errorFn(callbacks...), fnWithIgnore(fn, errors.IsNotFound, callbacks...))
}

func OnGetError(backoff wait.Backoff, fn func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, errorFn(callbacks...), fn)
}
