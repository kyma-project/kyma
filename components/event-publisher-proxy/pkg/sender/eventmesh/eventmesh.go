package eventmesh

import (
	"context"
	"io"
	"net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2event "github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/eventmesh"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/common"
)

var _ sender.GenericSender = &Sender{}

var (
	// additionalHeaders are the required headers by EMS for publish requests.
	// Any alteration or removal of those headers might cause publish requests to fail.
	additionalHeaders = http.Header{
		"qos":    []string{string(eventmesh.QosAtLeastOnce)},
		"Accept": []string{internal.ContentTypeApplicationJSON},
	}
)

// Sender is responsible for sending messages over HTTP.
type Sender struct {
	Client *http.Client
	Target string
}

func (s *Sender) URL() string {
	return s.Target
}

func (s *Sender) Checker() *health.ConfigurableChecker {
	return &health.ConfigurableChecker{}
}

func (s *Sender) Send(ctx context.Context, event *cev2event.Event) sender.PublishError {
	request, err := s.NewRequestWithTarget(ctx, s.Target)
	if err != nil {
		e := common.ErrInternalBackendError
		e.Wrap(err)
		return e
	}

	message := binding.ToMessage(event)
	defer func() { _ = message.Finish(nil) }()

	err = cloudevents.WriteRequestWithHeaders(ctx, message, request, additionalHeaders)
	if err != nil {
		e := common.ErrInternalBackendError
		e.Wrap(err)
		return e
	}

	resp, err := s.Client.Do(request)
	if err != nil {
		e := common.ErrInternalBackendError
		e.Wrap(err)
		return e
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return common.ErrInternalBackendError
	}
	return common.BackendPublishError{HttpCode: resp.StatusCode, Info: string(body)}
}

// NewSender returns a new Sender instance with the given target and client.
func NewSender(target string, client *http.Client) *Sender {
	return &Sender{Client: client, Target: target}
}

// NewRequestWithTarget returns a new HTTP POST request with the given context and target.
func (s *Sender) NewRequestWithTarget(ctx context.Context, target string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
}
