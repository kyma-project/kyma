package http

import (
	"context"

	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"
)

type WrappedCloudEventClient struct {
	underlying cloudevents.Client
	options    []retry.Option
}

type ResilientCloudEventClient interface {
	Send(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error)
}

func NewWrappedCloudEventClient(ceClient cloudevents.Client, opts ...retry.Option) *WrappedCloudEventClient {
	var client = &WrappedCloudEventClient{
		underlying: ceClient,
		options:    opts,
	}
	return client
}

func (c *WrappedCloudEventClient) Send(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error) {
	err = retry.Do(func() error {
		ct, evt, err = c.underlying.Send(ctx, event)
		return err
	}, c.options...)
	return ct, evt, err
}
