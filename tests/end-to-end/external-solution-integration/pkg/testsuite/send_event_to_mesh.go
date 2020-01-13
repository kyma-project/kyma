package testsuite

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// SendEvent is a step which sends an example event to the application gateway
type SendEventToMesh struct {
	state   SendEventState
	appName string
}

var _ step.Step = &SendEventToMesh{}

// NewSendEvent returns new SendEvent
func NewSendEventToMesh(appName string, state SendEventState) *SendEventToMesh {
	return &SendEventToMesh{state: state, appName: appName}
}

// Name returns name name of the step
func (s *SendEventToMesh) Name() string {
	return "Send Cloud Event to Mesh"
}

// Run executes the step
func (s *SendEventToMesh) Run() error {
	ctx := context.TODO()
	event := s.prepareEvent()

	_, _, err := s.state.GetEventSender().SendCloudEventToMesh(ctx, event)
	return err
}

func (s *SendEventToMesh) prepareEvent() cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	data := "some data"
	event.SetID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	event.SetType(example_schema.EventType)
	event.SetSource("some source")
	event.SetData(data)
	event.SetTime(time.Now())
	event.SetExtension("eventtypeversion", example_schema.EventVersion)

	return event
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToMesh) Cleanup() error {
	return nil
}
