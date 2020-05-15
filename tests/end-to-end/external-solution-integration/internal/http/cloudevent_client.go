package http

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	baseretry "github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
)

const retryAttemptsCount = 20
const retryDelay = 250 * time.Millisecond

var defaultOpts = []baseretry.Option{
	baseretry.Attempts(retryAttemptsCount),
	baseretry.Delay(retryDelay),
	baseretry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "WrappedCloudEventClient").Debugf("OnRetry: attempts: %d, error: %v", n, err)
	}),
}

type WrappedCloudEventClient struct {
	underlying cloudevents.Client
	options    []baseretry.Option
}

type ResilientCloudEventClient interface {
	Send(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error)
}

func NewWrappedCloudEventClient(ceClient cloudevents.Client, opts ...baseretry.Option) *WrappedCloudEventClient {
	var client = &WrappedCloudEventClient{
		underlying: ceClient,
		options:    append(defaultOpts, opts...),
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
