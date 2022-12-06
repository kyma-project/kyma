package sender

import (
	"context"
	"errors"

	"github.com/cloudevents/sdk-go/v2/event"
)

var (
	ErrInsufficientStorage   = errors.New("insufficient storage on backend")
	ErrBackendTargetNotFound = errors.New("publishing target on backend not found")
	ErrNoConnection          = errors.New("no connection to backend")
	ErrInternalBackendError  = errors.New("internal error on backend")
)

type GenericSender interface {
	Send(context.Context, *event.Event) (PublishResult, error)
	URL() string
}

type PublishResult interface {
	HTTPStatus() int
	ResponseBody() []byte
}
