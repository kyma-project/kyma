package controller

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	RequeueTime = 10 * time.Second
)

// ShouldRetryOn indicates if an error from the kubernetes client should be retried.
// Errors caused by a bad request or configuration should not be retried.
func ShouldRetryOn(err error) bool {
	return !errors.IsInvalid(err) &&
		!errors.IsNotAcceptable(err) &&
		!errors.IsUnsupportedMediaType(err) &&
		!errors.IsMethodNotSupported(err) &&
		!errors.IsBadRequest(err) &&
		!errors.IsUnauthorized(err) &&
		!errors.IsForbidden(err) &&
		!errors.IsNotFound(err)
}
