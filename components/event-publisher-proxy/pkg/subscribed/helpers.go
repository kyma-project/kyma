package subscribed

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
)

const (
	DefaultResyncPeriod = 10 * time.Second
)

var (
	GVR = schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
)

// ConvertRuntimeObjToSubscription converts a runtime.Object to a Subscription object by converting to Unstructed in between
func ConvertRuntimeObjToSubscription(sObj runtime.Object) (*eventingv1alpha1.Subscription, error) {
	sub := &eventingv1alpha1.Subscription{}
	if subUnstructured, ok := sObj.(*unstructured.Unstructured); ok {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(subUnstructured.Object, sub)
		if err != nil {
			return nil, err
		}
	}
	return sub, nil
}

// GenerateSubscriptionInfFactory generates DynamicSharedInformerFactory for Subscription
func GenerateSubscriptionInfFactory(k8sConfig *rest.Config) dynamicinformer.DynamicSharedInformerFactory {
	subDynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(subDynamicClient,
		DefaultResyncPeriod,
		v1.NamespaceAll,
		nil,
	)
	dFilteredSharedInfFactory.ForResource(GVR)
	return dFilteredSharedInfFactory
}

type waitForCacheSyncFunc func(stopCh <-chan struct{}) map[schema.GroupVersionResource]bool

// WaitForCacheSyncOrDie waits for the cache to sync. If sync fails everything stops
func WaitForCacheSyncOrDie(ctx context.Context, dc dynamicinformer.DynamicSharedInformerFactory) {
	dc.Start(ctx.Done())

	ctx, cancel := context.WithTimeout(context.Background(), DefaultResyncPeriod)
	defer cancel()

	err := hasSynced(ctx, dc.WaitForCacheSync)
	if err != nil {
		log.Fatalf("Failed to sync informer caches: %v", err)
	}
}

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

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-synced:
	}

	return nil
}

// ConvertEventsMapToSlice converts a map of Events to a slice of Events
func ConvertEventsMapToSlice(eventsMap map[Event]bool) []Event {
	result := make([]Event, 0)
	for k, _ := range eventsMap {
		result = append(result, k)
	}
	return result
}

// AddUniqueEventsToResult returns a map of unique Events which also contains the events eventsSubSet
func AddUniqueEventsToResult(eventsSubSet []Event, uniqEvents map[Event]bool) map[Event]bool {
	if len(uniqEvents) == 0 {
		uniqEvents = make(map[Event]bool)
	}
	for _, event := range eventsSubSet {
		if !uniqEvents[event] {
			uniqEvents[event] = true
		}
	}
	return uniqEvents
}

// FilterEventTypeVersions returns a slice of Events:
// 1. if the eventType matches the format: <eventTypePrefix><appName>.<event-name>.<version>
// E.g. sap.kyma.custom.varkes.order.created.v0
// 2. if the eventSource matches BEBNamespace name
func FilterEventTypeVersions(eventTypePrefix, bebNs, appName string, filters *eventingv1alpha1.BebFilters) []Event {
	events := make([]Event, 0)
	if filters == nil {
		return events
	}
	for _, filter := range filters.Filters {
		if filter == nil {
			continue
		}
		searchingForEventPrefix := strings.ToLower(fmt.Sprintf("%s.%s.", eventTypePrefix, appName))

		if filter.EventSource != nil && filter.EventType != nil {
			if strings.ToLower(filter.EventSource.Value) == strings.ToLower(bebNs) && strings.HasPrefix(filter.EventType.Value, searchingForEventPrefix) {
				eventTypeVersion := strings.ReplaceAll(filter.EventType.Value, searchingForEventPrefix, "")
				eventTypeVersionArr := strings.Split(eventTypeVersion, ".")
				version := eventTypeVersionArr[len(eventTypeVersionArr)-1]
				eventType := ""
				for i, part := range eventTypeVersionArr {
					if i == 0 {
						eventType = part
						continue
					}
					// Adding the segments till last but 1 as the last one is the version
					if i < (len(eventTypeVersionArr) - 1) {
						eventType = fmt.Sprintf("%s.%s", eventType, part)
					}
				}
				event := Event{
					Name:    eventType,
					Version: version,
				}
				events = append(events, event)
			}
		}
	}
	return events
}
