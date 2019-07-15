package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"time"
)

type SendEvent struct {
	state SendEventState
}

type SendEventState interface {
	GetEventSender() *testkit.EventSender
}

var _ step.Step = &SendEvent{}

func NewSendEvent(state SendEventState) *SendEvent {
	return &SendEvent{state: state}
}

func (s *SendEvent) Name() string {
	return "Send event"
}

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

func (s *SendEvent) Cleanup() error {
	return nil
}
