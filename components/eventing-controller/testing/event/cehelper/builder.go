package cehelper

import (
	"time"

	v2 "github.com/cloudevents/sdk-go/v2"
)

type EventOpt func(event *v2.Event) error

// NewEvent returns a CloudEvent build from passed EventOpts. This is only a helper that builds around existing
// CloudEvent functions.
func NewEvent(opts ...EventOpt) (*v2.Event, error) {
	e := v2.NewEvent()
	for _, o := range opts {
		if err := o(&e); err != nil {
			return nil, err
		}
	}
	return &e, nil
}

// WithType is a EventOpt to set the `ce-type` header of a CloudEvent. This will never return an error.
func WithType(_type string) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetType(_type)
		return nil
	}
}

// WithData is a EventOpt to set the data header of a CloudEvent. It accepts the data as a string and the encoding as an
// interface. A set of valid encodings can be found here:
// https://github.com/cloudevents/sdk-go/blob/2d3bb5b2c7e5eeab39e07bd47004cb17a83f0ca6/v2/alias.go#L58
func WithData(data string, encoding interface{}) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		if err := e.SetData(data, encoding); err != nil {
			return err
		}
		return nil
	}
}

// WithSubject is a EventOpt to set the `ce-subject` header of a CloudEvent. This will never return an error.
func WithSubject(subject string) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetSubject(subject)
		return nil
	}
}

// WithID is a EventOpt to set the `ce-id` header of a CloudEvent. This will never return an error.
func WithID(id string) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetID(id)
		return nil
	}
}

// WithSource is a EventOpt to set the `ce-source` header of a CloudEvent. This will never return an error.
func WithSource(source string) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetSource(source)
		return nil
	}
}

// WithSpecVersion is a EventOpt to set the `ce-specversion` header of a CloudEvent. This will never return an error.
func WithSpecVersion(version string) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetSpecVersion(version)
		return nil
	}
}

// WithTime is a EventOpt to set the `ce-time` header of a CloudEvent. This will never return an error.
func WithTime(timestamp time.Time) func(event *v2.Event) error {
	return func(e *v2.Event) error {
		e.SetTime(timestamp)
		return nil
	}
}
