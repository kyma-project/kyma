package retry

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	DefaultBackoff = retry.DefaultBackoff
)

func wrapErrorFunc(fn func() error, ignoreErr func(error) bool, callbacks ...func(...interface{})) func() error {
	return func() error {
		err := fn()
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

func shouldRetry(callbacks ...func(...interface{})) func(error) bool {
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

func OnCreateError(backoff wait.Backoff, errorFunc func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, shouldRetry(callbacks...), wrapErrorFunc(errorFunc, errors.IsAlreadyExists, callbacks...))
}

func OnDeleteError(backoff wait.Backoff, errorFunc func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, shouldRetry(callbacks...), wrapErrorFunc(errorFunc, errors.IsNotFound, callbacks...))
}

func OnGetError(backoff wait.Backoff, errorFunc func() error, callbacks ...func(...interface{})) error {
	return retry.OnError(backoff, shouldRetry(callbacks...), errorFunc)
}
