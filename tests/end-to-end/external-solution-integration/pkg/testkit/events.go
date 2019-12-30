package testkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/http"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/common/resilient"
	"github.com/pkg/errors"
)

type EventSender struct {
	httpClient resilient.HttpClient
	ceClient   http2.ResilientCloudEventClient
	domain     string
}

func NewEventSender(httpClient resilient.HttpClient, domain string, ceClient http2.ResilientCloudEventClient) *EventSender {
	return &EventSender{
		httpClient: httpClient,
		domain:     domain,
		ceClient:   ceClient,
	}
}

func (s *EventSender) SendEvent(appName string, event *ExampleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://gateway.%s/%s/v1/events", s.domain, appName)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("send event failed: %v\nrequest: %v\nresponse: %v", response.StatusCode, request, response)
	}

	return nil
}

func (s *EventSender) SendCloudEventToMesh(ctx context.Context, event cloudevents.Event) (ct context.Context, evt *cloudevents.Event, err error) {
	return s.ceClient.Send(ctx, event)
}