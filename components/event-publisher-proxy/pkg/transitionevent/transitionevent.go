package transitionevent

import (
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

var segmentNames = [4]string{
	"prefix", "application name", "event name", "version",
}

// TransitionEvent is a wrapper around a cloud event that allows to directly
// manipulate the segments of an event's type for the way types are used
// in kyma eventing.
type TransitionEvent struct {
	ce.Event
	prefix    string
	appName   string
	eventName string
	version   string
}

func NewTransitionEventFromCloudEvent(cloudEvent ce.Event, prefix, appName, eventName, version string) (*TransitionEvent) {
	transitionEvent := TransitionEvent{
		Event:     cloudEvent,
		prefix:    prefix,
		appName:   appName,
		eventName: eventName,
		version:   version,
	}

	transitionEvent.updateType()

	return &transitionEvent
}

func (e *TransitionEvent) Prefix() string {
	return e.prefix
}

func (e *TransitionEvent) SetPrefix(p string) {
	e.prefix = p
	e.updateType()
}

func (e *TransitionEvent) AppName() string {
	return e.appName
}

func (e *TransitionEvent) SetAppName(s string) {
	e.appName = s
	e.updateType()
}

func (e *TransitionEvent) EventName() string {
	return e.eventName
}

func (e *TransitionEvent) SetEventName(s string) {
	e.eventName = s
	e.updateType()
}

func (e *TransitionEvent) Version() string {
	return e.version
}

func (e *TransitionEvent) SetVersion(s string) {
	e.version = s
	e.updateType()
}

func (e *TransitionEvent) Type() string {
	return e.Event.Type()
}

func (e *TransitionEvent) updateType() {
	eventType := concatSegmentsWithDot(e.prefix, e.appName, e.eventName, e.version)
	e.Event.SetType(eventType)
}

// concatSegmentsWithDot takes an array of strings and concat them with a dot in between.
// ["a", "b", "c", "d", "e", "f"] would lead to "a.b.c.d.e.f".
func concatSegmentsWithDot(segments ...string) string {
	s := ""
	for _, segment := range segments {
		if s == "" {
			s = segment
			break
		}
		if segment != "" {
			s = fmt.Sprintf("%s.%s", s, segment)
		}
	}
	return s
}
