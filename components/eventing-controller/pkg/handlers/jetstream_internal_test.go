package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/mocks"
	testing2 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_getCallBack(t *testing.T) {
	// given

	const (
		givenSubKeyPrefix     = "givenSubKeyPrefix"
		givenSinkURL          = "http://localhorst:4444"
		givenSubscriptionName = "givenSubscriptionName"
	)
	sinks := sync.Map{}

	// store a sink for the given subscription
	sinks.Store(givenSubKeyPrefix, givenSinkURL)

	defaultLogger := fixtLogger(t)
	givenCloudEvent := fixtCloudEvent(t)
	messager := mocks.Messager{}

	// simulate that acking the message returns no error
	messager.On("Ack").Return(nil)

	// provide the underlying NATS message
	natsMessage := fixtNatsMessage(fixtCloudEventJson(t, givenCloudEvent), givenSubscriptionName)
	messager.On("Msg").Return(natsMessage)

	// mock the part where the cloud event is sent to the sink urlk part
	// NOTE: the mock was created using
	//       mockery --name "Client" --srcpkg github.com/cloudevents/sdk-go/v2/client
	cloudEventSender := mocks.Client{}
	cloudEventSender.On("Send", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		ensureSinkURL(ctx, t, givenSinkURL)

		// Ensure the passed event is correct.
		event := args.Get(1).(cloudevents.Event)
		assert.Equal(t, event.ID(), givenCloudEvent.ID(), "Error while matching the event IDs")
	})

	metricsCollector := mocks.CollectorInterface{}
	metricsCollector.On("RecordDeliveryPerSubscription", givenSubscriptionName, givenCloudEvent.Type(), givenSinkURL, http.StatusOK).Return()

	// create the object under test
	handler := JetStream{
		// TODO: change to pointer
		sinks:            sinks,
		client:           &cloudEventSender,
		logger:           defaultLogger,
		metricsCollector: &metricsCollector,
	}

	// when
	natsMsgHandler := handler.getCallback(givenSubKeyPrefix, givenSubscriptionName)
	natsMsgHandler(&messager)

	// then
	// expect that the event was sent
	cloudEventSender.AssertExpectations(t)

	// ensure the nats msg was acked
	messager.AssertExpectations(t)

	// ensure the metric was set
	metricsCollector.AssertExpectations(t)
}

// ensureSinkURL ensures that givenSinkURL is set as target URL in ctx.
func ensureSinkURL(ctx context.Context, t *testing.T, givenSinkURL string) {
	actualSink := cloudevents.TargetFromContext(ctx)
	assert.NotNil(t, actualSink, "Error while ensuring the sink is not nil")
	assert.Equal(t, actualSink.String(), givenSinkURL, "Error while matching the sink URLs.")
}

func fixtLogger(t *testing.T) *logger.Logger {
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	return defaultLogger
}

func fixtNatsMessage(cloudEvent []byte, subscriptionName string) *nats.Msg {
	return &nats.Msg{
		Subject: subscriptionName,
		Reply:   "some reply",
		Data:    cloudEvent,

		Sub: &nats.Subscription{
			Subject: subscriptionName,
			Queue:   "",
		},
	}
}

func fixtCloudEventJson(t *testing.T, cloudEvent *cloudevents.Event) []byte {
	cloudEventJson, err := json.Marshal(cloudEvent)
	require.Nil(t, err, "Error while serializing json")
	return cloudEventJson
}

func fixtCloudEvent(t *testing.T) *cloudevents.Event {
	cloudEvent, err := testing2.CloudEvent()
	require.Nil(t, err, "Error while building cloud event")
	return cloudEvent
}
