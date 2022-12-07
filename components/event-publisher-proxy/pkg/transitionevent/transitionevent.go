package transitionevent

import (
	"errors"
	"fmt"
	"strings"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

// TransitionEvent is a wrapper around a cloud event that allows to directly
// manipulate the segments of an event's type for the way types are used
// in kyma eventing.
type TransitionEvent struct {
	ce.Event

	// These are the segments of a cloud events Type:
	// 'prefix.appName.eventName.version'.
	prefix    string
	appName   string
	eventName string
	version   string
}

type Opt func(*TransitionEvent) error

func NewTransitionEvent(options ...Opt) (*TransitionEvent, error) {
	event := TransitionEvent{
		Event: ce.Event{},
	}

	// Apply all options.
	for _, o := range options {
		o(&event)
	}

	// If the type segments are not set yet, try to extract them from the
	// underlying cloud event.
	if event.areAllTypeSegmentsEmpty() {
		err := event.setTypeSegmentViaCloudEvent()
		if err != nil {
			return nil, err
		}
	}

	event.updateType()

	return &event, nil
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
	_type := concatSegmentsWithDot(e.prefix, e.appName, e.eventName, e.version)
	e.Event.SetType(_type)
}

// concatSegmentsWithDot takes an array of strings and concat them with a dot in
// between. For example ["a", "b", "c", "d", "e", "f"] would lead to
// "a.b.c.d.e.f".
func concatSegmentsWithDot(segments ...string) string {
	s := ""
	for _, segment := range segments {
		if s == "" {
			s = segment
			continue
		}
		if segment != "" {
			s = fmt.Sprintf("%s.%s", s, segment)
		}
	}
	return s
}

func (e *TransitionEvent) setTypeSegmentViaCloudEvent() error {
	prefix, appName, eventName, version, err := extractSegmentsFromEvent(&e.Event)
	if err != nil {
		return err
	}

	e.prefix = prefix
	e.appName = appName
	e.eventName = eventName
	e.version = version

	return nil
}

func (e *TransitionEvent) areAllTypeSegmentsEmpty() bool {
	return e.prefix == "" && e.appName == "" && e.eventName == "" && e.version == ""
}

func extractSegmentsFromEvent(event *ce.Event) (string, string, string, string, error) {
	segments := strings.Split(event.Type(), ".")
	length := len(segments)
	if len(segments) < 4 || checkForEmptySegments(segments) {
		return "", "", "", "", errors.New("invalid format")
	}

	version := segments[length-1]
	eventName := concatSegmentsWithDot(segments[length-2 : length-3]...)
	appName := segments[length-4]
	prefix := concatSegmentsWithDot(segments[length-5 : 0]...)

	return prefix, appName, eventName, version, nil
}

func checkForEmptySegments(segments []string) bool {
	for _, segment := range segments {
		if segment == "" {
			return true
		}
	}
	return false
}
