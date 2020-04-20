package subscribed

import (
	"context"
	"log"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	knv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/eventing/pkg/client/informers/externalversions"
	kneventinglister "knative.dev/eventing/pkg/client/listers/eventing/v1alpha1"
	"knative.dev/pkg/signals"
)

const (
	keySource           = "source"
	keyEventType        = "type"
	keyEventTypeVersion = "eventtypeversion"

	informerSyncTimeout = time.Second * 5
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
	triggerLister kneventinglister.TriggerLister
}

// waitForInformersSyncOrDie blocks until all informer caches are synced, or panics after a timeout.
func waitForInformersSyncOrDie(f externalversions.SharedInformerFactory) {
	log.Println("waiting for informers caches sync...")
	ctx, cancel := context.WithTimeout(context.Background(), informerSyncTimeout)
	defer cancel()

	err := hasSynced(ctx, f.WaitForCacheSync)
	if err != nil {
		log.Fatalf("Error waiting for caches sync: %s", err)
	}
}

type waitForCacheSyncFunc func(stopCh <-chan struct{}) map[reflect.Type]bool

// hasSynced blocks until the given informer sync waiting function completes. It returns an error if the passed context
// gets canceled.
func hasSynced(ctx context.Context, fn waitForCacheSyncFunc) error {
	// synced gets closed as soon as fn returns
	synced := make(chan struct{})

	// closing stopWait forces fn to return, which happens whenever ctx
	// gets canceled
	stopWait := make(chan struct{})
	defer close(stopWait)

	// close the synced channel if the `WaitForCacheSync()` finished the execution cleanly
	go func() {
		informersCacheSync := fn(stopWait)
		res := true
		for _, sync := range informersCacheSync {
			if !sync {
				res = false
			}
		}
		if res {
			close(synced)
		}
	}()

	// wait for closure of the goroutine or return an error if it timed out
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-synced:
	}

	return nil
}

//NewEventsClient function creates client for retrieving all active events
func NewEventsClient(client versioned.Interface) EventsClient {
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	lister := informerFactory.Eventing().V1alpha1().Triggers().Lister()
	ctx := signals.NewContext()
	informerFactory.Start(ctx.Done())
	waitForInformersSyncOrDie(informerFactory)

	return &eventsClient{
		triggerLister: lister,
	}
}

func (ec *eventsClient) GetSubscribedEvents(appName string) (Events, error) {
	activeEvents, err := ec.getEventList(appName)
	if err != nil {
		return Events{}, err
	}
	activeEvents = removeDuplicates(activeEvents)
	return Events{EventsInfo: activeEvents}, nil
}

func (ec *eventsClient) getEventList(appName string) ([]Event, error) {
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
		if source, ok := (*trigger.Spec.Filter.Attributes)[keySource]; ok && source == appName {
			event := Event{
				Name:    (*trigger.Spec.Filter.Attributes)[keyEventType],
				Version: (*trigger.Spec.Filter.Attributes)[keyEventTypeVersion],
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
