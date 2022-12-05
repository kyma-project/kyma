package transitionevent

import (
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

var segmentNames = [7]string{
	"prefix 1", "prefix 2", "prefix 3",
	"application name",
	"event name 1", "event name 2",
	"version",
}

// TransitionEvent is a wrapper around a cloud event that allows to directly
// manipulate the segments of an event's type for the way types are used
// in kyma eventing. 
type TransitionEvent struct {
	ce.Event
	prefix1, prefix2, prefix3 string
	appName                   string
	eventName1, eventName2    string
	version                   string
}

func NewFromCloudEvent(cloudEvent ce.Event,
	prefix1, prefix2, prefix3, 
	appName, 
	eventName1, eventName2, 
	version string) (*ce.Event, error) {
	// Check that all event type segments are not empty.
	for i, s := range []string{prefix1, prefix2, prefix3, appName, eventName1, eventName2, version} {
		if s == "" {
			return nil, fmt.Errorf("segment '%s' is empty string", segmentNames[i])
		}
	}
	
	// Create a new TransitionEvent.
	transitionEvent := TransitionEvent{
		Event: cloudEvent,
		prefix1: prefix1,
		prefix2: prefix2,
		prefix3: prefix3,
		appName: appName,
		eventName1: eventName1,
		eventName2: eventName2,
		version: version,
	}
	transitionEvent.updateType()

	return &transitionEvent.Event, nil
}

func (e *TransitionEvent) Prefix1() string {
	return e.prefix1
}

func (e *TransitionEvent) SetPrefix1(p string) {
	e.prefix1 = p
	e.updateType()
}

func (e *TransitionEvent) Prefix2() string {
	return e.prefix2
}

func (e *TransitionEvent) SetPrefix2(p string) {
	e.prefix2 = p
	e.updateType()
}

func (e *TransitionEvent) Prefix3() string {
	return e.prefix3
}

func (e *TransitionEvent) SetPrefix3(p string) {
	e.prefix3 = p
	e.updateType()
}

func (e *TransitionEvent) AppName() string {
	return e.appName
}

func (e *TransitionEvent) SetAppName(s string) {
	e.appName = s
	e.updateType()
}

func (e *TransitionEvent) EventName1() string {
	return e.eventName1
}

func (e *TransitionEvent) SetEventName1(s string) {
	e.eventName1 = s
	e.updateType()
}

func (e *TransitionEvent) EventName2() string {
	return e.eventName2
}

func (e *TransitionEvent) SetEventName2(s string) {
	e.eventName2 = s
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
	e.Event.SetType(concatSegmentsWithDot(
		e.prefix1, e.prefix2, e.prefix3,
		e.appName,
		e.eventName1, e.eventName2,
		e.version,
	))
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
