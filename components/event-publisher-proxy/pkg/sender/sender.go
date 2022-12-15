package sender

import (
	"context"
	"errors"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/builder"
)

var (
	ErrInsufficientStorage   = errors.New("insufficient storage on backend")
	ErrBackendTargetNotFound = errors.New("publishing target on backend not found")
	ErrNoConnection          = errors.New("no connection to backend")
	ErrInternalBackendError  = errors.New("internal error on backend")
)

type GenericSender interface {
	Send(context.Context, *builder.Event) (PublishResult, error)
	URL() string
}

type PublishResult interface {
	HTTPStatus() int
	ResponseBody() []byte
}
