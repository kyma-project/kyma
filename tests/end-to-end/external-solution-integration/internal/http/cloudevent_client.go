package http

import (
	"context"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"

	custom_retry "github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"
)

type WrappedCloudEventClient struct {
	underlying cloudevents.Client
	options    []custom_retry.Option
}

type ResilientCloudEventClient interface {
	Send(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error)
}

func NewWrappedCloudEventClient(ceClient cloudevents.Client, opts ...custom_retry.Option) *WrappedCloudEventClient {
	var client = &WrappedCloudEventClient{
		underlying: ceClient,
		options:    opts,
	}
	return client
}

func (c *WrappedCloudEventClient) Send(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error) {
	err = retry.WithCustomOpts(func() error {
		ct, evt, err = c.underlying.Send(ctx, event)
		return err
	}, c.options...)
	return ct, evt, err
}
