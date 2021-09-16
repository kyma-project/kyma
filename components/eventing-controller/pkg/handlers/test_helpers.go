package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	cev2nats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kyma-project/kyma/components/eventing-controller/testing"

	"time"

	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func SendEventToNATS(natsClient *Nats, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := eventingtesting.OrderCreatedEventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := NewNatsMessagePayload(data, "id", eventingtesting.EventSource, eventTime, eventType)
	return natsClient.connection.Publish(eventType, []byte(sampleEvent))
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}

func SendBinaryCloudEventToNATS(natsClient *Nats, subject string) error {
	// create a CE binary-mode http request
	body := testing.CloudEventData
	headers := testing.GetBinaryMessageHeaders()
	req, err := http.NewRequest(http.MethodPost, "dummy", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header[k] = v
	}
	// convert  to the CE Event
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	message := cev2http.NewMessageFromHttpRequest(req)
	defer func() { _ = message.Finish(nil) }()
	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return err
	}
	// get a CE sender for the embedded NATS using CE-SDK
	natsOpts := cev2nats.NatsOptions()
	url := natsClient.config.Url
	sender, err := cev2nats.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := cev2.NewClient(sender)
	if err != nil {
		return err
	}
	// force binary binding and send the event to NATS using CE-SDK
	ctx = binding.WithForceBinary(ctx)
	if err := client.Send(ctx, *event); err != nil {
		return err
	}
	return nil
}

func SendStructuredCloudEventToNATS(natsClient *Nats, subject string) error {
	// create a CE structured-mode http request
	body := testing.StructuredCloudEvent
	headers := testing.GetStructuredMessageHeaders()
	req, err := http.NewRequest(http.MethodPost, "dummy", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header[k] = v
	}
	// convert to CE Event
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	message := cev2http.NewMessageFromHttpRequest(req)
	defer func() { _ = message.Finish(nil) }()
	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return err
	}
	// get a CE sender for the embedded NATS
	natsOpts := cev2nats.NatsOptions()
	url := natsClient.config.Url
	sender, err := cev2nats.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := cev2.NewClient(sender)
	if err != nil {
		return err
	}
	// force structured binding and send the event to NATS
	ctx = binding.WithForceStructured(ctx)
	if err := client.Send(ctx, *event); err != nil {
		return err
	}
	return nil
}
