package common

import (
	"fmt"
	"strings"
	"sync"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
)

type EventDetails struct {
	eventType        string
	eventTypeVersion string
	source           *source
}

type source struct {
	sourceEnvironment string
	sourceNamespace   string
	sourceType        string
}

func FromPublishRequest(r *api.PublishRequest) *EventDetails {
	sourceStruct := &source{
		sourceEnvironment: r.Source.SourceEnvironment,
		sourceNamespace:   r.Source.SourceNamespace,
		sourceType:        r.Source.SourceType,
	}
	return &EventDetails{
		eventType:        r.EventType,
		eventTypeVersion: r.EventTypeVersion,
		source:           sourceStruct,
	}
}

//Replace all occurrences of the `.` with `\.` as well as `\` with `\\`
func escapePeriodsAndBackSlashes(in *string) string {
	s := strings.Replace(*in, `\`, `\\`, -1)
	return strings.Replace(s, `.`, `\.`, -1)
}

/*
Encode formats the event details into a NATS streaming compliant subject name literal.
Encoded subject is constructed by using the Period (`.`) character be added between tokens as a delimeter.
Period character in a token literal will be escaped with a forward slash (`\.`), ex: `env.prod` will be `env\.prod`.
Return value is astring literal composed of the event details tokens as: `sourceEnvironment + sourceNamespace + sourceType + eventType + eventTypeVersion`.
*/
func (e *EventDetails) Encode() string {
	return fmt.Sprintf(`%s.%s.%s.%s.%s`,
		escapePeriodsAndBackSlashes(&e.source.sourceEnvironment),
		escapePeriodsAndBackSlashes(&e.source.sourceNamespace),
		escapePeriodsAndBackSlashes(&e.source.sourceType),
		escapePeriodsAndBackSlashes(&e.eventType),
		escapePeriodsAndBackSlashes(&e.eventTypeVersion))
}

func FromSubscriptionSpec(s v1alpha1.SubscriptionSpec) *EventDetails {
	sourceStruct := &source{
		sourceEnvironment: s.Source.SourceEnvironment,
		sourceNamespace:   s.Source.SourceNamespace,
		sourceType:        s.Source.SourceType,
	}
	return &EventDetails{
		eventType:        s.EventType,
		eventTypeVersion: s.EventTypeVersion,
		source:           sourceStruct,
	}
}

type StatusReady struct {
	mu    sync.RWMutex
	ready bool
}

func (s *StatusReady) SetReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ready {
		s.ready = true
		return true
	}
	return false
}

func (s *StatusReady) SetNotReady() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ready {
		s.ready = false
	}
}
