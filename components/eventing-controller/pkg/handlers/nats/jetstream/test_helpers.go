package jetstream

import (
	"bytes"
	"context"
	"net/http"
	"time"

	nats2 "github.com/cloudevents/sdk-go/protocol/nats/v2"
	v2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"

	testing2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/testing"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func SendEventToJetStream(jsClient *JetStream, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := evtesting.OrderCreatedEventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := testing2.NewNatsMessagePayload(data, "id", evtesting.EventSource, eventTime, eventType)
	return jsClient.conn.Publish(jsClient.GetJetstreamSubject(eventType), []byte(sampleEvent))
}

func SendEventToJetStreamOnEventType(jsClient *JetStream, eventType string, data string) error {
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := testing2.NewNatsMessagePayload(data, "id", evtesting.EventSource, eventTime, eventType)
	return jsClient.conn.Publish(jsClient.GetJetstreamSubject(eventType), []byte(sampleEvent))
}

func SendBinaryCloudEventToJetStream(jetStreamClient *JetStream, subject, eventData string) error {
	// create a CE binary-mode http request
	headers := evtesting.GetBinaryMessageHeaders()
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
	natsOpts := nats2.NatsOptions()
	url := jetStreamClient.Config.URL
	sender, err := nats2.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := v2.NewClient(sender)
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

func SendStructuredCloudEventToJetStream(jetStreamClient *JetStream, subject, eventData string) error {
	// create a CE structured-mode http request
	headers := evtesting.GetStructuredMessageHeaders()
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
	natsOpts := nats2.NatsOptions()
	url := jetStreamClient.Config.URL
	sender, err := nats2.NewSender(url, subject, natsOpts)
	if err != nil {
		return nil
	}
	client, err := v2.NewClient(sender)
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
