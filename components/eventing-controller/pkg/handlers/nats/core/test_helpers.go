package core

import (
	"bytes"
	"context"
	"net/http"
	"time"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"

	testing2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/testing"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func SendEventToNATS(natsClient *Nats, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := eventingtesting.OrderCreatedEventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := testing2.NewNatsMessagePayload(data, "id", eventingtesting.EventSource, eventTime, eventType)
	return natsClient.connection.Publish(eventType, []byte(sampleEvent))
}

func SendEventToNATSOnEventType(natsClient *Nats, eventType string, data string) error {
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := testing2.NewNatsMessagePayload(data, "id", eventingtesting.EventSource, eventTime, eventType)
	return natsClient.connection.Publish(eventType, []byte(sampleEvent))
}

func SendBinaryCloudEventToNATS(natsClient *Nats, subject, eventData string) error {
	// create a CE binary-mode http request
	headers := eventingtesting.GetBinaryMessageHeaders()
	req, err := http.NewRequest(http.MethodPost, "dummy", bytes.NewBuffer([]byte(eventData)))
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
	natsOpts := cenats.NatsOptions()
	url := natsClient.Config.URL
	sender, err := cenats.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := ce.NewClient(sender)
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

func SendStructuredCloudEventToNATS(natsClient *Nats, subject, eventData string) error {
	// create a CE structured-mode http request
	headers := eventingtesting.GetStructuredMessageHeaders()
	req, err := http.NewRequest(http.MethodPost, "dummy", bytes.NewBuffer([]byte(eventData)))
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
	natsOpts := cenats.NatsOptions()
	url := natsClient.Config.URL
	sender, err := cenats.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := ce.NewClient(sender)
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
