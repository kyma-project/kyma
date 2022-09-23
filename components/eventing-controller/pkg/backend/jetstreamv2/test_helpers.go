package jetstreamv2

import (
	"bytes"
	"context"
	nats2 "github.com/cloudevents/sdk-go/protocol/nats/v2"
	v2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"net/http"
	"time"
)

func sendEventToJetStream(jsClient *JetStream, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := evtesting.OrderCreatedEventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := natstesting.NewNatsMessagePayload(data, "id", evtesting.EventSource, eventTime, eventType)
	return jsClient.conn.Publish(getJetStreamSubject(eventType), []byte(sampleEvent))
}

func sendEventToJetStreamOnEventType(jsClient *JetStream, eventType string, data string) error {
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := natstesting.NewNatsMessagePayload(data, "id", evtesting.EventSource, eventTime, eventType)
	return jsClient.conn.Publish(getJetStreamSubject(eventType), []byte(sampleEvent))
}

func sendCloudEventToJetStream(jetStreamClient *JetStream, subject, eventData, cetype string) error {
	// create a CE http request
	var headers http.Header
	if cetype == types.ContentModeBinary {
		headers = evtesting.GetBinaryMessageHeaders()
	} else {
		headers = evtesting.GetStructuredMessageHeaders()
	}
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
	if cetype == types.ContentModeBinary {
		ctx = binding.WithForceBinary(ctx)
	} else {
		ctx = binding.WithForceStructured(ctx)
	}
	if err := client.Send(ctx, *event); err != nil {
		return err
	}
	return nil
}
