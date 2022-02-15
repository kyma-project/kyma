package retry

import (
	goerrors "errors"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	DefaultBackoff = retry.DefaultBackoff
	ErrInvalidFunc = goerrors.New("invalid function")
)

func fnWithIgnore(fn func() error, ignoreErr func(error) bool, log *logrus.Entry) func() error {
	return func() error {
		if fn == nil {
			return ErrInvalidFunc
		}
		err := fn()
		if ignoreErr == nil {
			return err
		}
		if ignoreErr(err) {
			log.Infof("ignoring: %s", err)
			return nil
		}
		return err
	}
}

func errorFn(log *logrus.Entry) func(error) bool {
	return func(err error) bool {
		if errors.IsTimeout(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) {
			log.Infof("retrying due to: %s", err)
			return true
		}
		return false
	}
}

func WithIgnoreOnNotFound(backoff wait.Backoff, fn func() error, log *logrus.Entry) error {
	return retry.OnError(backoff, errorFn(log), fnWithIgnore(fn, errors.IsNotFound, log))
}

func WithIgnoreOnConflict(backoff wait.Backoff, fn func() error, log *logrus.Entry) error {
	return retry.OnError(backoff, errorFn(log), fnWithIgnore(fn, errors.IsConflict, log))
}

func OnTimeout(backoff wait.Backoff, fn func() error, log *logrus.Entry) error {
	return retry.OnError(backoff, errorFn(log), fn)
}
