package subscribed

import (
	subscriptions "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

//EventsClient interface
type EventsClient interface {
	GetSubscribedEvents(appName string) (Events, error)
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
	subscriptionsClient *subscriptions.Clientset
	namespacesClient    core.NamespaceInterface
}

//NewEventsClient function creates client for retrieving all active events
func NewEventsClient() (EventsClient, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *rest.Config) (EventsClient, error) {
	subscriptionsClient, err := subscriptions.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	coreClient, err := core.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	namespaces := coreClient.Namespaces()

	return &eventsClient{
		subscriptionsClient: subscriptionsClient,
		namespacesClient:    namespaces,
	}, nil
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
	subscriptionList, e := ec.subscriptionsClient.EventingV1alpha1().Subscriptions(namespace).List(meta.ListOptions{})

	if e != nil {
		return nil, e
	}

	events := make([]Event, 0)

	for _, subscription := range subscriptionList.Items {
		if subscription.SourceID == appName {
			event := Event{Name: subscription.EventType, Version: subscription.EventTypeVersion}
			events = append(events, event)
		}
	}

	return events, nil
}

func (ec *eventsClient) getAllNamespaces() ([]string, error) {
	namespaceList, e := ec.namespacesClient.List(meta.ListOptions{})

	if e != nil {
		return nil, e
	}

	var namespaces []string

	for _, namespace := range namespaceList.Items {
		namespaces = append(namespaces, namespace.Name)
	}

	return namespaces, nil
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
