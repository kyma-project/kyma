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
	payload string
}

var _ step.Step = &SendEventToMesh{}

// NewSendEvent returns new SendEvent
func NewSendEventToMesh(appName, payload string, state SendEventState) *SendEventToMesh {
	return &SendEventToMesh{state: state, appName: appName, payload: payload}
}

// Name returns name name of the step
func (s *SendEventToMesh) Name() string {
	return "Send Cloud Event to Mesh"
}

// Run executes the step
func (s *SendEventToMesh) Run() error {
	ctx := context.TODO()
	event, err := s.prepareEvent()
	if err != nil {
		return err
	}

	_, _, err = s.state.GetEventSender().SendCloudEventToMesh(ctx, event)
	return err
}

func (s *SendEventToMesh) prepareEvent() (cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
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

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToMesh) Cleanup() error {
	return nil
}
