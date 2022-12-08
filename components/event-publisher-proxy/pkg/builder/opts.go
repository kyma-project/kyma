package builder

import (
	"net/http"

	cesdk "github.com/cloudevents/sdk-go/v2"
	ce "github.com/cloudevents/sdk-go/v2/event"

	et "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
)

// WithCloudEvent will set the TransitionEvent's event as the given one.
func WithCloudEvent(cloudEvent *ce.Event) func(*Event) error {
	return func(t *Event) error {
		t.Event = *cloudEvent
		return nil
	}
}

// WithCleaner will use the given cleaner to clean up the cloud event's type.
// Remember to put this option after other options that set the type in the
// first place.
func WithCleaner(cleaner et.Cleaner) func(*Event) error {
	return func(t *Event) error {
		cleanType, err := cleaner.Clean(t.Event.Type())

		t.SetType(cleanType)

		return err
	}
}

// WithPrefix will set the prefix segment.
func WithPrefix(prefix string) func(*Event) error {
	return func(t *Event) error {
		t.prefix = consolidateToMaxNumberOfSegments(3, prefix)
		return nil
	}
}

// WithApp will set the appName segment.
func WithApp(app string) func(*Event) error {
	return func(t *Event) error {
		t.app = consolidateToMaxNumberOfSegments(1, app)
		return nil
	}
}

// WithEventName will set the eventName segment.
func WithName(name string) func(*Event) error {
	return func(t *Event) error {
		t.name = consolidateToMaxNumberOfSegments(2, name)
		return nil
	}
}

// WithVersion will set the version segment.
func WithVersion(version string) func(*Event) error {
	return func(t *Event) error {
		t.version = consolidateToMaxNumberOfSegments(1, version)
		return nil
	}
}

// WithCloudEventFromRequest will try to extract a cloud event from a request
// and use it as the TransferEvent's underlying event.
func WithCloudEventFromRequest(request *http.Request) func(*Event) error {
	return func(event *Event) error {
		cloudEvent, err := cesdk.NewEventFromHTTPRequest(request)
		if err != nil {
			return err
		}

		err = cloudEvent.Validate()
		if err != nil {
			return err
		}

		event.Event = *cloudEvent
		return nil
	}
}

// WithRemoveNonAlphanumericsFromType will remove all non-alphanumeric characters (but ".") from the type.
func WithRemoveNonAlphanumericsFromType() func(*Event) error {
	return func(event *Event) error {
		event.prefix = removeNonAlphanumeric(event.prefix)
		event.version = removeNonAlphanumeric(event.version)
		event.name = removeNonAlphanumeric(event.name)
		event.app = removeNonAlphanumeric(event.app)
		return nil
	}
}
