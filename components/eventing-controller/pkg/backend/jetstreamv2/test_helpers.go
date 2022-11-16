package jetstreamv2

import (
	"bytes"
	"context"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats-server/v2/server"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	nats2 "github.com/cloudevents/sdk-go/protocol/nats/v2"
	v2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	DefaultStreamName    = "kyma"
	DefaultMaxReconnects = 10
	DefaultMaxInFlights  = 10
)

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	jsBackend  *JetStream
	logger     *logger.Logger
	natsServer *server.Server
	jsClient   *jetStreamClient
	natsConfig env.NatsConfig
	cleaner    cleaner.Cleaner
	natsPort   int
}

func SendEventToJetStream(jsClient *JetStream, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := evtestingv2.OrderCreatedCleanEvent
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := natstesting.NewNatsMessagePayload(data, "id", evtestingv2.EventSourceClean, eventTime, eventType)
	return jsClient.Conn.Publish(jsClient.GetJetStreamSubject(evtestingv2.EventSourceClean,
		eventType,
		v1alpha2.TypeMatchingStandard,
	), []byte(sampleEvent))
}

func SendCloudEventToJetStream(jetStreamClient *JetStream, subject, eventData, cetype string) error {
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
	if validateErr := event.Validate(); validateErr != nil {
		return validateErr
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

func AddJSCleanEventTypesToStatus(sub *v1alpha2.Subscription, cleaner cleaner.Cleaner) {
	cleanEventType := GetCleanEventTypes(sub, cleaner)
	sub.Status.Types = cleanEventType
}
