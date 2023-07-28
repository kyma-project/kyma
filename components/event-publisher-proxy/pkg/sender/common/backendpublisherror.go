package common

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
)

var _ sender.PublishError = &BackendPublishError{}

//nolint:lll //reads better this way
var (
	ErrInsufficientStorage    = BackendPublishError{HTTPCode: http.StatusInsufficientStorage, Info: "insufficient storage on backend"}
	ErrBackendTargetNotFound  = BackendPublishError{HTTPCode: http.StatusNotFound, Info: "publishing target on backend not found"}
	ErrClientNoConnection     = BackendPublishError{HTTPCode: http.StatusBadGateway, Info: "no connection to backend"}
	ErrInternalBackendError   = BackendPublishError{HTTPCode: http.StatusInternalServerError, Info: "internal error on backend"}
	ErrClientConversionFailed = BackendPublishError{HTTPCode: http.StatusBadRequest, Info: "conversion to target format failed"}
)

type BackendPublishError struct {
	HTTPCode int
	Info     string
	err      error
}

func (e BackendPublishError) Error() string {
	return e.Info
}

func (e *BackendPublishError) Unwrap() error {
	return e.err
}

func (e *BackendPublishError) Wrap(wrappedError error) {
	e.err = wrappedError
}

func (e BackendPublishError) Code() int {
	if e.HTTPCode == 0 {
		return http.StatusInternalServerError
	}
	return e.HTTPCode
}

func (e BackendPublishError) Message() string {
	return e.Info
}

func (e *BackendPublishError) Is(target error) bool {
	t, ok := target.(*BackendPublishError) //nolint:errorlint //we dont want to check the error chain here
	if !ok {
		return false
	}
	return (e.HTTPCode == t.HTTPCode) && (e.Info == t.Info)
}
