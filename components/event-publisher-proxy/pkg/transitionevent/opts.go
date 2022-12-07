package transitionevent

import (
	ce "github.com/cloudevents/sdk-go/v2/event"
	et "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
)

func WithCloudEvent(cloudEvent *ce.Event) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.Event = *cloudEvent
		return nil
	}
}

func WithCleaner(cleaner et.Cleaner) func(*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		cleanType, err := cleaner.Clean(t.Event.Type())

		t.SetType(cleanType)
		
		return err
	}
}

func WithPrefix(prefix string) func (*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.prefix = prefix
		return nil
	}
}

func WithAppName(appName string) func (*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.appName = appName
		return nil
	}
}

func WithEventName(eventName string) func (*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.eventName = eventName
		return nil
	}
}

func WithVersion(version string) func (*TransitionEvent) error {
	return func(t *TransitionEvent) error {
		t.version = version
		return nil
	}
}