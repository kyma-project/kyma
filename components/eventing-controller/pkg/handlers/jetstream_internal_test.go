package handlers

import (
	"encoding/json"
	"sync"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"
	testing2 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_getCallBack(t *testing.T) {
	// given

	const (
		subKeyPrefix     = "subKeyPrefix"
		sinkURL          = "some url"
		subscriptionName = "subscriptionName"
	)
	sinks := sync.Map{}

	// store a sink for the given subscription
	sinks.Store(subKeyPrefix, sinkURL)

	defaultLogger := fixtLogger(t)
	metricsCollector := metrics.NewCollector()
	cloudEvent := fixtCloudEvent(t)
	m := mocks.Messager{}
	m.On("Ack").Return(nil)
	natsMessage := fixtNatsMessage(fixtCloudEventJson(t, cloudEvent), subscriptionName)
	m.On("Msg").Return(natsMessage)
	// natsMessage := fixtNatsMessage(fixtCloudEventJson(t, cloudEvent), subscriptionName)
	// TODO: how to ack the message ?
	// require.NoError(t, natsMessage.Ack(), "Error while acking the NATS message")

	// mock the part where the cloud event is sent to the sink urlk part
	// NOTE: the mock was created using
	//       mockery --name "Client" --srcpkg github.com/cloudevents/sdk-go/v2/client
	cloudEventSender := mocks.Client{}
	cloudEventSender.On("Send", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		// TODO(nils): ensure that the context has the correct url
		// Unfortunately the context key is not available for us to be used :/
		//
		// Ensure the sink URL is correctly set in the context.
		// ctx := args.Get(0).(context.Context)
		// sink := ctx.Value(cloudevents.ContextWithTarget).(string)
		// assert.Equal(t, sink, sinkURL, "Error while matching the sink URLs.")

		// Ensure the passed event is correct.
		event := args.Get(1).(cloudevents.Event)
		assert.Equal(t, event.ID(), cloudEvent.ID(), "Error while matching the event IDs")
	})
	// create the object under test
	handler := JetStream{
		// TODO: change to pointer
		sinks:            sinks,
		client:           &cloudEventSender,
		logger:           defaultLogger,
		metricsCollector: metricsCollector,
	}

	// when
	natsMsgHandler := handler.getCallback(subKeyPrefix, subscriptionName)
	natsMsgHandler(&m)

	// then
	// expect that the event was sent
	cloudEventSender.AssertExpectations(t)
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
