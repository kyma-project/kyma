package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"time"
)

// SendEvent is a step which sends example event to the application gateway
type SendEvent struct {
	state SendEventState
}

// SendEventState represents SendEvent dependencies
type SendEventState interface {
	GetEventSender() *testkit.EventSender
}

var _ step.Step = &SendEvent{}

// NewSendEvent returns new SendEvent
func NewSendEvent(state SendEventState) *SendEvent {
	return &SendEvent{state: state}
}

// Name returns name name of the step
func (s *SendEvent) Name() string {
	return "Send event"
}

// Run executes the step
func (s *SendEvent) Run() error {
	event := s.prepareEvent()
	return s.state.GetEventSender().SendEvent(consts.AppName, event)
}

func (s *SendEvent) prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        consts.EventType,
		EventTypeVersion: consts.EventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             "some data",
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *SendEvent) Cleanup() error {
	return nil
}
