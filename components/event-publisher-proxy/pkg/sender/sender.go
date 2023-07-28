package sender

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

type GenericSender interface {
	Send(context.Context, *event.Event) PublishError
	URL() string
}

type PublishError interface {
	error
	Code() int
	Message() string
}
