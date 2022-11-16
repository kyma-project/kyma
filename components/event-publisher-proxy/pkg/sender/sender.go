package sender

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

type GenericSender interface {
	Send(context.Context, *event.Event) (PublishResult, error)
	URL() string
}

type PublishResult interface {
	HTTPStatus() int
	ResponseBody() []byte
}
