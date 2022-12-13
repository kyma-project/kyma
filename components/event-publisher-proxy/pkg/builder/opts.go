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

// WithOriginalType sets the OriginalType.
func WithOriginalType(originalType string) func(*Event) error {
	return func(e *Event) error {
		e.originalType = originalType
		return nil
	}
}

// WithCleaner will use the given cleaner to clean up the cloud event's type. Remember to put this option after other
// options that set the type in the first place. If ..
// TODO
func WithCleaner(cleaner et.Cleaner) func(*Event) error {
	return func(e *Event) error {
		cleanType, err := cleaner.Clean(e.Event.Type())

		e.SetType(cleanType)

		return err
	}
}

// WithPrefix will set the prefix segment.
func WithPrefix(prefix string) func(*Event) error {
	return func(e *Event) error {
		e.prefix = consolidateToMaxNumberOfSegments(3, prefix)
		return nil
	}
}

// WithApp will set the appName segment.
func WithApp(app string) func(*Event) error {
	return func(e *Event) error {
		e.app = consolidateToMaxNumberOfSegments(1, app)
		return nil
	}
}

// WithEventName will set the eventName segment.
func WithName(name string) func(*Event) error {
	return func(e *Event) error {
		e.name = consolidateToMaxNumberOfSegments(2, name)
		return nil
	}
}

// WithVersion will set the version segment.
func WithVersion(version string) func(*Event) error {
	return func(e *Event) error {
		e.version = consolidateToMaxNumberOfSegments(1, version)
		return nil
	}
}

// WithCloudEventFromRequest will try to extract a cloud event from a request and use it as the TransferEvent's
// underlying event.
func WithCloudEventFromRequest(request *http.Request) func(*Event) error {
	return func(e *Event) error {
		cloudEvent, err := cesdk.NewEventFromHTTPRequest(request)
		if err != nil {
			return err
		}

		err = cloudEvent.Validate()
		if err != nil {
			return err
		}

		e.Event = *cloudEvent
		return nil
	}
}

// WithRemoveNonAlphanumericsFromType will remove all non-alphanumeric characters (but ".") from the type.
func WithRemoveNonAlphanumericsFromType() func(*Event) error {
	return func(e *Event) error {
		e.prefix = removeNonAlphanumeric(e.prefix)
		e.version = removeNonAlphanumeric(e.version)
		e.name = removeNonAlphanumeric(e.name)
		e.app = removeNonAlphanumeric(e.app)
		return nil
	}
}

func WithEventSource(source string) func(*Event) error {
	return func(e *Event) error {
		e.Event.SetSource(source)
		return nil
	}
}

func WithEventExtension(name string, obj interface{}) func(*Event) error {
	return func(e *Event) error {
		e.Event.SetExtension(name, obj)
		return nil
	}
}

func WithEventDataContentType(contentType string) func(*Event) error {
	return func(e *Event) error {
		e.Event.SetDataContentType(contentType)
		return nil
	}
}
