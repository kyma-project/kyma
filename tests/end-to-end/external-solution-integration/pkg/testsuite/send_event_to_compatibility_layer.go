package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// SendEventToCompatibilityLayer is a step which sends example event to the application gateway
type SendEventToCompatibilityLayer struct {
	testkit.SendEvent
}

var _ step.Step = &SendEventToCompatibilityLayer{}

// NewSendEventToCompatibilityLayer returns new SendEventToCompatibilityLayer
func NewSendEventToCompatibilityLayer(appName, payload string, state testkit.SendEventState) *SendEventToCompatibilityLayer {
	return &SendEventToCompatibilityLayer{testkit.SendEvent{State: state, AppName: appName, Payload: payload}}
}

// Name returns name name of the step
func (s *SendEventToCompatibilityLayer) Name() string {
	return "Send event to compatibility layer"
}

// Run executes the step
func (s *SendEventToCompatibilityLayer) Run() error {
	event := s.prepareEvent()
	return s.State.GetEventSender().SendEventToCompatibilityLayer(s.AppName, event)
}

func (s *SendEventToCompatibilityLayer) prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        example_schema.EventType,
		EventTypeVersion: example_schema.EventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             s.Payload,
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEventToCompatibilityLayer) Cleanup() error {
	return nil
}
