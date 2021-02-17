package subscribed

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

var (
	GVR = schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
)

// ConvertRuntimeObjToSubscription converts a runtime.Object to a Subscription object by converting to unstructured in between
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
		informers.DefaultResyncPeriod,
		v1.NamespaceAll,
		nil,
	)
	dFilteredSharedInfFactory.ForResource(GVR)
	return dFilteredSharedInfFactory
}

// ConvertEventsMapToSlice converts a map of Events to a slice of Events
func ConvertEventsMapToSlice(eventsMap map[Event]bool) []Event {
	result := make([]Event, 0)
	for k := range eventsMap {
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
