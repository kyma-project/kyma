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

// SendEventToMeshAndCheckEventId is a step which sends an event and checks if the correct EventId has been received
type SendEventToMeshAndCheckEventId struct {
	state       SendEventState
	appName     string
	payload     string
	testService *testkit.TestService
	retryOpts   []retrygo.Option
}

var counterMesh = 0
var _ step.Step = &SendEventToMeshAndCheckEventId{}

// NewSendEventToMeshAndCheckEventId returns new SendEventToMeshAndCheckEventId
func NewSendEventToMeshAndCheckEventId(appName, payload string, state SendEventState, testService *testkit.TestService,
	opts ...retrygo.Option) *SendEventToMeshAndCheckEventId {
	return &SendEventToMeshAndCheckEventId{
		appName:     appName,
		payload:     payload,
		state:       state,
		testService: testService,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *SendEventToMeshAndCheckEventId) Name() string {
	return "Send event to mesh and check event id"
}

// Run executes the step
func (s *SendEventToMeshAndCheckEventId) Run() error {
	const basicId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa"
	eventId := fmt.Sprint(basicId, counterMesh)

	err := s.sendEventToMesh(eventId)
	if err != nil {
		return err
	}

	err = s.checkEventId(eventId)
	if err != nil {
		counterMesh++
		return s.testService.CheckAllReceivedEvents()
	}

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToMeshAndCheckEventId) Cleanup() error {
	return nil
}

func (s *SendEventToMeshAndCheckEventId) checkEventId(eventId string) error {
	err := retry.Do(func() error {
		return s.testService.CheckEventId(eventId)
	}, s.retryOpts...)

	return err
}

func (s *SendEventToMeshAndCheckEventId) sendEventToMesh(eventId string) error {
	ctx := context.TODO()
	event, err := s.prepareEvent(eventId)
	if err != nil {
		return err
	}

	_, _, err = s.state.GetEventSender().SendCloudEventToMesh(ctx, event)
	logrus.WithField("component", "SendEventToMesh").
		Debugf("SendCloudEventToMesh: eventID: %v; error: %v", event.ID(), err)
	return err
}

func (s *SendEventToMeshAndCheckEventId) prepareEvent(eventId string) (cloudevents.Event, error) {
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
