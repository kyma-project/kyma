package subscribed

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/typed/eventing.kyma-project.io/v1alpha1"
	log "github.com/sirupsen/logrus"
	coretypes "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	knv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	kneventingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	kneventinglister "knative.dev/eventing/pkg/client/listers/eventing/v1alpha1"
)

//EventsClient interface
type EventsClient interface {
	GetSubscribedEvents(appName string) (Events, error)
}

//SubscriptionsGetter interface
type SubscriptionsGetter interface {
	Subscriptions(namespace string) v1alpha1.SubscriptionInterface
}

type KnativeTriggerGetter interface {
	Triggers(namespace string) kneventingv1alpha1.TriggerInterface
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
	triggerLister       kneventinglister.TriggerLister
}

//NewEventsClient function creates client for retrieving all active events
func NewEventsClient(subscriptionsClient SubscriptionsGetter, namespacesClient NamespacesClient, triggerLister kneventinglister.TriggerLister) EventsClient {
	return &eventsClient{
		subscriptionsClient: subscriptionsClient,
		namespacesClient:    namespacesClient,
		triggerLister:       triggerLister,
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
		log.Infof("error retrieving triggers : %+v", err)
		return nil, err
	}

	bt, err := json.Marshal(triggerList)
	log.Infof("trigger list: %+v", string(bt))
	log.Infof("marshal error: %+v", err)
	return getEventsFromTriggers(triggerList, appName), nil
}

func getEventsFromTriggers(triggerList []*knv1alpha1.Trigger, appName string) []Event {
	events := make([]Event, 0, len(triggerList))

	for _, trigger := range triggerList {
		log.Infof("trigger: %+v", trigger)
		if trigger == nil || trigger.Spec.Filter == nil || trigger.Spec.Filter.Attributes == nil {
			continue
		}
		if source, ok := (*trigger.Spec.Filter.Attributes)["source"]; ok && source == appName { //TODO(marcobebway) evaluate possible nil dereference
			log.Infof("source name: %s. appname: %s", source, appName) ///TODO(marcobebway) remove this
			event := Event{
				Name:    (*trigger.Spec.Filter.Attributes)["type"],             //TODO(marcobebway) evaluate possible nil dereference
				Version: (*trigger.Spec.Filter.Attributes)["eventtypeversion"], //TODO(marcobebway) evaluate possible nil dereference
			}
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
