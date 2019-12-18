package subscribed

import (
	"github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"

	eventtypes "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"

	coretypes "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//EventsClient interface
type EventsClient interface {
	GetSubscribedEvents(appName string) (Events, error)
}

//SubscriptionsGetter interface
type SubscriptionsGetter interface {
	Subscriptions(namespace string) v1alpha1.SubscriptionInterface
}

//NamespacesClient interface
type NamespacesClient interface {
	List(opts meta.ListOptions) (*coretypes.NamespaceList, error)
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
	subscriptionsClient SubscriptionsGetter
	namespacesClient    NamespacesClient
}

//NewEventsClient function creates client for retrieving all active events
func NewEventsClient(subscriptionsClient SubscriptionsGetter, namespacesClient NamespacesClient) EventsClient {

	return &eventsClient{
		subscriptionsClient: subscriptionsClient,
		namespacesClient:    namespacesClient,
	}
}

func (ec *eventsClient) GetSubscribedEvents(appName string) (Events, error) {
	var activeEvents []Event

	namespaces, e := ec.getAllNamespaces()

	if e != nil {
		return Events{}, e
	}

	for _, namespace := range namespaces {
		events, err := ec.getEventsForNamespace(appName, namespace)
		if err != nil {
			return Events{}, err
		}
		activeEvents = append(activeEvents, events...)
	}

	activeEvents = removeDuplicates(activeEvents)

	return Events{EventsInfo: activeEvents}, nil
}

func (ec *eventsClient) getEventsForNamespace(appName, namespace string) ([]Event, error) {
	subscriptionList, e := ec.subscriptionsClient.Subscriptions(namespace).List(meta.ListOptions{})

	if e != nil {
		return nil, e
	}

	return getEventsFromSubscriptions(subscriptionList, appName), nil
}

func getEventsFromSubscriptions(subscriptionList *eventtypes.SubscriptionList, appName string) []Event {
	events := make([]Event, 0)

	for _, subscription := range subscriptionList.Items {
		if subscription.SourceID == appName {
			event := Event{Name: subscription.EventType, Version: subscription.EventTypeVersion}
			events = append(events, event)
		}
	}

	return events
}

func (ec *eventsClient) getAllNamespaces() ([]string, error) {
	namespaceList, e := ec.namespacesClient.List(meta.ListOptions{})

	if e != nil {
		return nil, e
	}

	return extractNamespacesNames(namespaceList), nil
}

func extractNamespacesNames(namespaceList *coretypes.NamespaceList) []string {
	var namespaces []string

	for _, namespace := range namespaceList.Items {
		namespaces = append(namespaces, namespace.Name)
	}

	return namespaces
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
