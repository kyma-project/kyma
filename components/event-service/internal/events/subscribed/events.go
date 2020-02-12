package subscribed

import (
	"k8s.io/apimachinery/pkg/labels"
	knv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	kneventingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	kneventinglister "knative.dev/eventing/pkg/client/listers/eventing/v1alpha1"
)

//EventsClient interface
type EventsClient interface {
	GetSubscribedEvents(appName string) (Events, error)
}

type KnativeTriggerGetter interface {
	Triggers(namespace string) kneventingv1alpha1.TriggerInterface
}

//Events represents collection of all events with subscriptions
type Events struct {
	EventsInfo []Event `json:"eventsInfo"`
}

//Event represents basic information about event
type Event struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type eventsClient struct {
	triggerLister kneventinglister.TriggerLister
}

//NewEventsClient function creates client for retrieving all active events
func NewEventsClient(triggerLister kneventinglister.TriggerLister) EventsClient {
	return &eventsClient{
		triggerLister: triggerLister,
	}
}

func (ec *eventsClient) GetSubscribedEvents(appName string) (Events, error) {
	activeEvents, err := ec.getKnativeTriggers(appName)
	if err != nil {
		return Events{}, err
	}
	activeEvents = removeDuplicates(activeEvents)
	return Events{EventsInfo: activeEvents}, nil
}

func (ec *eventsClient) getKnativeTriggers(appName string) ([]Event, error) {
	triggerList, err := ec.triggerLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	return getEventsFromTriggers(triggerList, appName), nil
}

func getEventsFromTriggers(triggerList []*knv1alpha1.Trigger, appName string) []Event {
	events := make([]Event, 0, len(triggerList))

	for _, trigger := range triggerList {
		if trigger == nil {
			continue
		}
		if trigger.Spec.Filter == nil {
			continue
		}
		if trigger.Spec.Filter.Attributes == nil {
			continue
		}
		if source, ok := (*trigger.Spec.Filter.Attributes)["source"]; ok && source == appName {
			event := Event{
				Name:    (*trigger.Spec.Filter.Attributes)["type"],
				Version: (*trigger.Spec.Filter.Attributes)["eventtypeversion"],
			}
			events = append(events, event)
		}
	}
	return events
}

func removeDuplicates(events []Event) []Event {
	keys := make(map[Event]bool)
	list := make([]Event, 0)
	for _, event := range events {
		if _, value := keys[event]; !value {
			keys[event] = true
			list = append(list, event)
		}
	}
	return list
}
