package builder

import (
	"errors"
	"strings"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

// Event is a wrapper around a CloudEvent that allows to directly manipulate
// the segments of an event type for the way types are used
// in kyma eventing.
type Event struct {
	ce.Event

	// These are the segments of a CloudEvent's Type:
	// 'prefix.app.name.version'.
	prefix  string
	app     string
	name    string
	version string
}

// Opt is a type for options to build an Event.
type Opt func(*Event) error

// NewEvent will take options to create an Event.
func NewEvent(options ...Opt) (*Event, error) {
	event := Event{
		Event: ce.Event{},
	}

	// Apply all options.
	for _, o := range options {
		if err := o(&event); err != nil {
			return nil, err
		}
	}

	// If the type segments are not set yet try to extract them from the
	// underlying CloudEvent. If any segment --expect prefix-- is empty,
	// this will fail with an error.
	if event.areAllTypeSegmentsEmpty() {
		err := event.setTypeSegmentsViaCloudEvent()
		if err != nil {
			return nil, err
		}
	}

	event.updateType()

	return &event, nil
}

// Prefix only returns the prefix of the EventType.
func (e *Event) Prefix() string {
	return e.prefix
}

// SetPrefix sets the prefix and updates the Type accordingly.
func (e *Event) SetPrefix(p string) {
	e.prefix = p
	e.updateType()
}

// IsPrefixEmpty checks if the prefix is empty and returns a corresponding bool.
func (e *Event) IsPrefixEmpty() bool {
	return e.prefix != ""
}

// AppName only returns the prefix of the EventType.
func (e *Event) AppName() string {
	return e.app
}

// SetAppName sets the appName and updates the Type accordingly.
func (e *Event) SetAppName(s string) {
	e.app = s
	e.updateType()
}

// EventName only returns the prefix of the EventType.
func (e *Event) EventName() string {
	return e.name
}

// SetEventName sets the eventName and updates the Type accordingly.
func (e *Event) SetEventName(s string) {
	e.name = s
	e.updateType()
}

// Version only returns the version of the EventType.
func (e *Event) Version() string {
	return e.version
}

// SetVersion sets the version and updates the Type accordingly.
func (e *Event) SetVersion(s string) {
	e.version = s
	e.updateType()
}

// Type returns EventType.
func (e *Event) Type() string {
	return e.Event.Type()
}

func (e *Event) updateType() {
	_type := concatSegmentsWithDot(e.prefix, e.app, e.name, e.version)
	e.Event.SetType(_type)
}

// setTypeSegmentsViaCloudEvent will set the four type segments by trying to
// extract them from the underlying cloud event.
func (e *Event) setTypeSegmentsViaCloudEvent() error {
	prefix, appName, eventName, version, err := extractSegmentsFromEvent(&e.Event)
	if err != nil {
		return err
	}

	e.prefix = prefix
	e.app = appName
	e.name = eventName
	e.version = version

	return nil
}

// areAllTypeSegmentsEmpty checks if all type segments are empty and returns a
// corresponding bool.
func (e *Event) areAllTypeSegmentsEmpty() bool {
	return e.prefix == "" && e.app == "" && e.name == "" && e.version == ""
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

// isAnySegmentEmpty checks if any segment is empty and returns a corresponding
// bool.
func isAnySegmentEmpty(segments []string) bool {
	for _, segment := range segments {
		if segment == "" {
			return true
		}
	}
	return false
}
