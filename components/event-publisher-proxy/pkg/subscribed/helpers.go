package subscribed

import (
	"fmt"
	"strings"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
)

var (
	GVR = schema.GroupVersionResource{
		Version:  eventingv1alpha2.GroupVersion.Version,
		Group:    eventingv1alpha2.GroupVersion.Group,
		Resource: "subscriptions",
	}
	GVRV1alpha1 = schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
)

// ConvertRuntimeObjToSubscriptionV1alpha1 converts a runtime.Object to a v1alpha1 version of Subscription object
// by converting to unstructured in between
func ConvertRuntimeObjToSubscriptionV1alpha1(sObj runtime.Object) (*eventingv1alpha1.Subscription, error) {
	sub := &eventingv1alpha1.Subscription{}
	if subUnstructured, ok := sObj.(*unstructured.Unstructured); ok {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(subUnstructured.Object, sub)
		if err != nil {
			return nil, err
		}
	}
	return sub, nil
}

// ConvertRuntimeObjToSubscription converts a runtime.Object to a Subscription object by converting to unstructured in between
func ConvertRuntimeObjToSubscription(sObj runtime.Object) (*eventingv1alpha2.Subscription, error) {
	sub := &eventingv1alpha2.Subscription{}
	if subUnstructured, ok := sObj.(*unstructured.Unstructured); ok {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(subUnstructured.Object, sub)
		if err != nil {
			return nil, err
		}
	}
	return sub, nil
}

// GenerateSubscriptionInfFactory generates DynamicSharedInformerFactory for Subscription
func GenerateSubscriptionInfFactory(k8sConfig *rest.Config, enableNewCrd bool) dynamicinformer.DynamicSharedInformerFactory {
	subDynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(subDynamicClient,
		informers.DefaultResyncPeriod,
		v1.NamespaceAll,
		nil,
	)
	if enableNewCrd {
		dFilteredSharedInfFactory.ForResource(GVR)
	} else {
		dFilteredSharedInfFactory.ForResource(GVRV1alpha1)
	}
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
// if the event source matches the appName for typeMatching standard
// if the <eventTypePrefix>.<appName> is present in the eventType for typeMatching exact
func FilterEventTypeVersions(eventTypePrefix, appName string, subscription *eventingv1alpha2.Subscription) []Event {
	events := make([]Event, 0)
	prefixAndAppName := fmt.Sprintf("%s.%s.", eventTypePrefix, appName)

	for _, eventType := range subscription.Spec.Types {
		if subscription.Spec.TypeMatching == eventingv1alpha2.TypeMatchingExact {
			// in case of type matching exact, we have app name as a part of event type
			if strings.HasPrefix(eventType, prefixAndAppName) {
				eventTypeVersion := strings.ReplaceAll(eventType, prefixAndAppName, "")
				event := buildEvent(eventTypeVersion)
				events = append(events, event)
			}
		} else {
			// in case of type matching standard, the source must be app name
			if appName == subscription.Spec.Source {
				event := buildEvent(eventType)
				events = append(events, event)
			}
		}
	}
	return events
}

// it receives event and type version, e.g. order.created.v1 and returns {Name: order.created, Version: v1}
func buildEvent(eventTypeAndVersion string) Event {
	lastDotIndex := strings.LastIndex(eventTypeAndVersion, ".")
	eventName := eventTypeAndVersion[:lastDotIndex]
	eventVersion := eventTypeAndVersion[lastDotIndex+1:]
	return Event{
		Name:    eventName,
		Version: eventVersion,
	}
}

// FilterEventTypeVersionsV1alpha1 returns a slice of Events for v1alpha1 version of Subscription resource:
// 1. if the eventType matches the format: <eventTypePrefix><appName>.<event-name>.<version>
// E.g. sap.kyma.custom.varkes.order.created.v0
// 2. if the eventSource matches BEBNamespace name
func FilterEventTypeVersionsV1alpha1(eventTypePrefix, bebNs, appName string, filters *eventingv1alpha1.BEBFilters) []Event {
	events := make([]Event, 0)
	if filters == nil {
		return events
	}
	for _, filter := range filters.Filters {
		if filter == nil {
			continue
		}

		var prefix string
		if len(strings.TrimSpace(eventTypePrefix)) == 0 {
			prefix = strings.ToLower(fmt.Sprintf("%s.", appName))
		} else {
			prefix = strings.ToLower(fmt.Sprintf("%s.%s.", eventTypePrefix, appName))
		}

		if filter.EventSource != nil && filter.EventType != nil {
			// TODO revisit the filtration logic as part of https://github.com/kyma-project/kyma/issues/10761

			// filter by event-source if exists
			if len(strings.TrimSpace(filter.EventSource.Value)) > 0 && !strings.EqualFold(filter.EventSource.Value, bebNs) {
				continue
			}

			if strings.HasPrefix(filter.EventType.Value, prefix) {
				eventTypeVersion := strings.ReplaceAll(filter.EventType.Value, prefix, "")
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
