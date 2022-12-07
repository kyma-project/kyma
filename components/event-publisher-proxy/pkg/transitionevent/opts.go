package transitionevent

import (
	"context"
	"net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	ce "github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	et "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
)

// WithCloudEvent will set the TransitionEvent's event as the given one.
func WithCloudEvent(cloudEvent *ce.Event) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.Event = *cloudEvent
		return nil
	}
}

// WithCleaner will use the given cleaner to clean up the cloud event's type.
// Remember to put this option after other options that set the type in the
// first place.
func WithCleaner(cleaner et.Cleaner) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		cleanType, err := cleaner.Clean(t.Event.Type())

		t.SetType(cleanType)

		return err
	}
}

// WithPrefix will set the prefix segment.
func WithPrefix(prefix string) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.prefix = prefix
		return nil
	}
}

// WithAppName will set the appName segment.
func WithAppName(appName string) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.appName = appName
		return nil
	}
}

// WithEventName will set the eventName segment.
func WithEventName(eventName string) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.eventName = eventName
		return nil
	}
}

// WithVersion will set the version segment.
func WithVersion(version string) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.version = version
		return nil
	}
}

// WithCloudEventFromRequest will try to extract a cloud event from a request
// and use it as the TransferEvent's underlying event.
func WithCloudEventFromRequest(request *http.Request) func(*TransitionEvent) error {
	return func(transitionEvent *TransitionEvent) error {
		message := cehttp.NewMessageFromHttpRequest(request)
		defer func() { _ = message.Finish(nil) }()

		cloudEvent, err := binding.ToEvent(context.Background(), message)
		if err != nil {
			return err
		}

		err = cloudEvent.Validate()
		if err != nil {
			return err
		}

		transitionEvent.Event = *cloudEvent
		return nil
	}
}
