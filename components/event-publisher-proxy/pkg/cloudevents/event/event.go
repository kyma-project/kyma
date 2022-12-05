package event

import (
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2/event"
)

// Event is a wrapper around a cloud event that allows to directly
// manipulate the segments of an event's type for the way types are used
// in 
type Event struct {
	ce.Event
	prefix1, prefix2, prefix3 string
	appName                   string
	eventName1, eventName2    string
	version                   string
}

func (e *Event) Prefix1() string {
	return e.prefix1
}

func (e *Event) SetPrefix1(p string) {
	e.prefix1 = p
	e.updateType()
}

func (e *Event) Prefix2() string {
	return e.prefix2
}

func (e *Event) SetPrefix2(p string) {
	e.prefix2 = p
	e.updateType()
}

func (e *Event) Prefix3() string {
	return e.prefix3
}

func (e *Event) SetPrefix3(p string) {
	e.prefix3 = p
	e.updateType()
}

func (e *Event) AppName() string {
	return e.appName
}

func (e *Event) SetAppName(s string) {
	e.appName = s
	e.updateType()
}

func (e *Event) EventName1() string {
	return e.eventName1
}

func (e *Event) SetEventName1(s string) {
	e.eventName1 = s
	e.updateType()
}

func (e *Event) EventName2() string {
	return e.eventName2
}

func (e *Event) SetEventName2(s string) {
	e.eventName2 = s
	e.updateType()
}

func (e *Event) Version() string {
	return e.version
}

func (e *Event) SetVersion(s string) {
	e.version = s
	e.updateType()
}

func (e *Event) Type() string {
	return e.Event.Type()
}

func (e *Event) updateType() {
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
