package testsuite

import (
	"fmt"
	"time"

	retrygo "github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/sirupsen/logrus"
)

// SendEventToCompatibilityLayerAndCheckEventId is a step which sends an event and checks if the correct EventId has been received
type SendEventToCompatibilityLayerAndCheckEventId struct {
	state       SendEventState
	appName     string
	payload     string
	testService *testkit.TestService
	retryOpts   []retrygo.Option
}

var counterLayer = 0
var _ step.Step = &SendEventToCompatibilityLayerAndCheckEventId{}

// NewSendEventToCompatibilityLayerAndCheckEventId returns new SendEventToCompatibilityLayerAndCheckEventId
func NewSendEventToCompatibilityLayerAndCheckEventId(appName, payload string, state SendEventState, testService *testkit.TestService,
	opts ...retrygo.Option) *SendEventToCompatibilityLayerAndCheckEventId {
	return &SendEventToCompatibilityLayerAndCheckEventId{
		appName:     appName,
		payload:     payload,
		state:       state,
		testService: testService,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *SendEventToCompatibilityLayerAndCheckEventId) Name() string {
	return "Send event to compatibility layer and check event id"
}

// Run executes the step
func (s *SendEventToCompatibilityLayerAndCheckEventId) Run() error {
	const basicId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa"
	eventId := fmt.Sprint(basicId, counterLayer)

	err := s.sendEventToCompatibilityLayer(eventId)
	if err != nil {
		return err
	}

	err = s.checkEventId(eventId)
	if err != nil {
		counterLayer++
		return s.testService.CheckAllReceivedEvents()
	}

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToCompatibilityLayerAndCheckEventId) Cleanup() error {
	return nil
}

func (s *SendEventToCompatibilityLayerAndCheckEventId) checkEventId(eventId string) error {
	err := retry.Do(func() error {
		return s.testService.CheckEventId(eventId)
	}, s.retryOpts...)

	return err
}

func (s *SendEventToCompatibilityLayerAndCheckEventId) sendEventToCompatibilityLayer(eventId string) error {
	event := s.prepareEvent(eventId)
	err := s.state.GetEventSender().SendEventToCompatibilityLayer(s.appName, event)
	logrus.WithField("component", "SendEventToCompatibilityLayer").Debugf("SendCloudEventToCompatibilityLayer: eventID: %v; error: %v", eventId, err)

	return err
}

func (s *SendEventToCompatibilityLayerAndCheckEventId) prepareEvent(eventId string) *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        example_schema.EventType,
		EventTypeVersion: example_schema.EventVersion,
		EventID:          eventId,
		EventTime:        time.Now(),
		Data:             s.payload,
	}
}
