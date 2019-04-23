package registered

import (
	subscriptions "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

//EventsClient interface
type EventsClient interface {
	GetActiveEvents(appName string) (ActiveEvents, error)
}

//ActiveEvents represents collection of all events with subscriptions
type ActiveEvents struct {
	Events []string `json:"events"`
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

func (ec *eventsClient) GetActiveEvents(appName string) (ActiveEvents, error) {
	var activeEvents []string

	namespaces, e := ec.getAllNamespaces()

	if e != nil {
		return ActiveEvents{}, e
	}

	for _, namespace := range namespaces {
		events, err := ec.getEventsForNamespace(appName, namespace)
		if err != nil {
			return ActiveEvents{}, err
		}
		activeEvents = append(activeEvents, events...)
	}

	return ActiveEvents{Events: activeEvents}, nil
}

func (ec *eventsClient) getEventsForNamespace(appName, namespace string) ([]string, error) {
	subscriptionList, e := ec.subscriptionsClient.EventingV1alpha1().Subscriptions(namespace).List(meta.ListOptions{})

	if e != nil {
		return nil, e
	}

	events := make([]string, 0)

	for _, subscription := range subscriptionList.Items {
		if subscription.SourceID == appName {
			events = append(events, subscription.EventType)
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
