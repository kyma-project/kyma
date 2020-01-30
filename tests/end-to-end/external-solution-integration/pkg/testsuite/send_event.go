package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// SendEvent is a step which sends example event to the application gateway
type SendEvent struct {
	state   SendEventState
	appName string
	payload string
}

// SendEventState represents SendEvent dependencies
type SendEventState interface {
	GetEventSender() *testkit.EventSender
}

var _ step.Step = &SendEvent{}

// NewSendEvent returns new SendEvent
func NewSendEvent(appName, payload string, state SendEventState) *SendEvent {
	return &SendEvent{state: state, appName: appName, payload: payload}
}

// Name returns name name of the step
func (s *SendEvent) Name() string {
	return "Send event"
}

// Run executes the step
func (s *SendEvent) Run() error {
	event := s.prepareEvent()
	return s.state.GetEventSender().SendEvent(s.appName, event)
}

func (s *SendEvent) prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        example_schema.EventType,
		EventTypeVersion: example_schema.EventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             s.payload,
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEvent) Cleanup() error {
	return nil
}
