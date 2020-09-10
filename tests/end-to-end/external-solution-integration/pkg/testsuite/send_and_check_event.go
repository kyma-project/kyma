package testsuite

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"

	retrygo "github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/sirupsen/logrus"
)

// SendEventAndCheckId is a step which sends and event and checks if the correct EventId has been received
type SendEventAndCheckId struct {
	state       SendEventState
	appName     string
	payload     string
	testService *testkit.TestService
	retryOpts   []retrygo.Option
}

var _ step.Step = &SendEventAndCheckId{}

// NewSendEventAndCheckId returns new SendEventAndCheckId
func NewSendEventAndCheckId(appName, payload string, state SendEventState, testService *testkit.TestService,
	opts ...retrygo.Option) *SendEventAndCheckId {
	return &SendEventAndCheckId{
		appName:     appName,
		payload:     payload,
		state:       state,
		testService: testService,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *SendEventAndCheckId) Name() string {
	return "Send event and check event id"
}

// Run executes the step
func (s *SendEventAndCheckId) Run() error {
	const basicId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa"
	counter := 0
	eventId := fmt.Sprint(basicId, counter)
	// send event to mesh
	err := s.sendEventToMesh(eventId)

	if err != nil {
		return err
	}

	// check if event with correct id has been received
	err = s.checkEventId(eventId)

	// if not, check all received ce
	if err != nil {
		return s.testService.CheckAllReceivedEvents()
	}

	// increase counter
	counter++

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventAndCheckId) Cleanup() error {
	return nil
}

func (s *SendEventAndCheckId) checkEventId(eventId string) error {
	err := retry.Do(func() error {
		return s.testService.CheckEventId(eventId)
	}, s.retryOpts...)

	return err
}

func (s *SendEventAndCheckId) sendEventToMesh(eventId string) error {
	ctx := context.TODO()
	event, err := s.prepareEvent(eventId)
	if err != nil {
		return err
	}

	_, _, err = s.state.GetEventSender().SendCloudEventToMesh(ctx, event)
	logrus.WithField("component", "SendEventToMesh").Debugf("SendCloudEventToMesh: eventID: %v; error: %v", event.ID(), err)
	return err
}

func (s *SendEventAndCheckId) prepareEvent(eventId string) (cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(eventId)
	event.SetType(example_schema.EventType)
	event.SetSource("some source")
	// TODO(k15r): infer mime type automatically
	event.SetDataContentType("text/plain")
	if err := event.SetData(s.payload); err != nil {
		return event, err
	}

	event.SetTime(time.Now())
	event.SetExtension("eventtypeversion", example_schema.EventVersion)

	return event, nil
}
