package builder

import (
	"errors"
	"strings"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

// Event is a wrapper around a CloudEvent that allows to directly manipulate the segments of type of an event.
// in kyma eventing.
type Event struct {
	cloudEvent   *ce.Event

	
	// These are the segments of a CloudEvent's Type: "prefix.app.name.version".
	prefix       string
	app          string
	name         string
	version      string

	// This is the original type of the event before any changes were done to it. 
	originalType string
}

// Opt is a type for options to build an Event.
type Opt func(*Event) error

// NewEvent will take options to create an Event.
func NewEvent(options ...Opt) (*Event, error) {
	event := Event{
		cloudEvent: &ce.Event{},
	}

	// Apply all options.
	for _, o := range options {
		if err := o(&event); err != nil {
			return nil, err
		}
	}

	// If the type segments are not set yet try to extract them from the underlying CloudEvent. If any segment
	// --expect the prefix-- is empty, this will fail with an error.
	if event.areAllTypeSegmentsEmpty() {
		err := event.setTypeSegmentsViaCloudEvent()
		if err != nil {
			return nil, err
		}
	}

	event.updateType()

	return &event, nil
}


func (e *Event) CloudEvent() *ce.Event {
    return e.cloudEvent
}

func (e *Event) SetCloudEvent(cloudEvent *ce.Event) error {
    e.cloudEvent = cloudEvent
    err := e.setTypeSegmentsViaCloudEvent()
    return err
} 

// Prefix only returns the prefix segment of the Type.
func (e *Event) Prefix() string {
	return e.prefix
}

// SetPrefix sets the prefix segment of the Type and updates the Type accordingly.
func (e *Event) SetPrefix(p string) {
	e.prefix = p
	e.updateType()
}

// IsPrefixEmpty checks if the prefix is empty and returns a corresponding bool.
func (e *Event) IsPrefixEmpty() bool {
	return e.prefix != ""
}

// Name only returns the app segment of the Type.
func (e *Event) Name() string {
	return e.app
}

// SetApp sets the app segment of the Type and updates the Type accordingly.
func (e *Event) SetApp(s string) {
	e.app = s
	e.updateType()
}

// EventName only returns the name segment of the Type.
func (e *Event) EventName() string {
	return e.name
}

// SetName sets the name and updates the Type accordingly.
func (e *Event) SetName(s string) {
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
	return e.cloudEvent.Type()
}


// TODO
func (e *Event) SetOriginalEventType(s string) {
    e.originalType = s
}

//TODO
//TODO can originalType be publis instead
func (e *Event) OriginalEventType() string {
    return e.originalType
}

func (e *Event) updateType() {
	_type := concatSegmentsWithDot(e.prefix, e.app, e.name, e.version)
	e.cloudEvent.SetType(_type)
}

// setTypeSegmentsViaCloudEvent will set the four type segments by trying to
// extract them from the underlying cloud event.
func (e *Event) setTypeSegmentsViaCloudEvent() error {
	prefix, appName, eventName, version, err := extractSegmentsFromEvent(e.cloudEvent)
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
