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

// NewTransitionEvent will take options to create a TransitionEvent.
func NewTransitionEvent(options ...Opt) (*TransitionEvent, error) {
	event := TransitionEvent{
		Event: ce.Event{},
	}

	// Apply all options.
	for _, o := range options {
		if err := o(&event); err != nil {
			return nil, err
		}
	}

	// If the type segments are not set yet try to extract them from the
	// underlying cloud event. If any segment expect prefix is empty this will
	// fail with an error.
	if event.areAllTypeSegmentsEmpty() {
		err := event.setTypeSegmentsViaCloudEvent()
		if err != nil {
			return nil, err
		}
	}

	event.updateType()

	return &event, nil
}

// Prefix only returns the prefix of the event's type that usually has the
// structure 'prefix.appName.eventName.version'.
func (e *TransitionEvent) Prefix() string {
	return e.prefix
}

func (e *TransitionEvent) SetPrefix(p string) {
	e.prefix = p
	e.updateType()
}

func (e *TransitionEvent) IsPrefixEmpty() bool {
	return e.prefix != ""
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

// setTypeSegmentsViaCloudEvent will set the four type segments by trying to
// extract them from the underlying cloud event.
func (e *TransitionEvent) setTypeSegmentsViaCloudEvent() error {
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

// areAllTypeSegmentsEmpty checks if all type segments are empty and returns a
// corresponding bool.
func (e *TransitionEvent) areAllTypeSegmentsEmpty() bool {
	return e.prefix == "" && e.appName == "" && e.eventName == "" && e.version == ""
}

// extractSegmentsFromEvent tries extract the event type's segments from a given
// cloud event. It will return an error if not at least version, event name and
// app name can be extracted while the prefix stays optional.
func extractSegmentsFromEvent(event *ce.Event) (string, string, string, string, error) {
	segments := strings.Split(event.Type(), ".")
	length := len(segments)
	if len(segments) < 4 || isAnySegmentEmpty(segments) {
		return "", "", "", "", errors.New("invalid format")
	}

	version := segments[length-1]
	eventName := concatSegmentsWithDot(segments[length-2 : length-3]...)
	appName := segments[length-4]
	prefix := concatSegmentsWithDot(segments[length-5 : 0]...)

	return prefix, appName, eventName, version, nil
}

// isAnySegmentEmpty will return false if any segment is empty.
func isAnySegmentEmpty(segments []string) bool {
	for _, segment := range segments {
		if segment == "" {
			return true
		}
	}
	return false
}
