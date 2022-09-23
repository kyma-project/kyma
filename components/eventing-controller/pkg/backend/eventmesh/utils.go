package eventmesh

import (
	"fmt"
	"strings"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// GetEventMeshSubject appends the prefix to subject.
func GetEventMeshSubject(source, subject, eventMeshPrefix string) string {
	return fmt.Sprintf("%s.%s.%s", eventMeshPrefix, source, subject)
}

// IsEventTypeSegmentsOverLimit checks if the number of segments in event type
// do not exceed EventMeshTypeSegmentsLimit
func IsEventTypeSegmentsOverLimit(eventType string) bool {
	segments := strings.Split(eventType, ".")
	return len(segments) > EventMeshTypeSegmentsLimit
}

func updateHashesInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshLocalSubscription *types.Subscription, eventMeshServerSubscription *types.Subscription) error {
	if err := setEventMeshLocalSubHashInStatus(kymaSubscription, eventMeshLocalSubscription); err != nil {
		return err
	}
	if err := setEventMeshServerSubHashInStatus(kymaSubscription, eventMeshServerSubscription); err != nil {
		return err
	}
	return nil
}

// setEventMeshLocalSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func setEventMeshLocalSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) error {
	// generate hash
	newHash, err := backendutils.GetHash(eventMeshSubscription)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.Ev2hash = newHash
	return nil
}

// setEventMeshServerSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func setEventMeshServerSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) error {
	// clean up the server sub object from extra info
	cleanedEventMeshSub := backendutils.GetCleanedEventMeshSubscription(eventMeshSubscription)
	// generate hash
	newHash, err := backendutils.GetHash(cleanedEventMeshSub)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.Emshash = newHash
	return nil
}

func statusCleanEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.EventType {
	var cleanEventTypes []eventingv1alpha2.EventType
	for _, i := range typeInfos {
		cleanEventTypes = append(cleanEventTypes, eventingv1alpha2.EventType{OriginalType: i.OriginalType, CleanType: i.CleanType})
	}
	return cleanEventTypes
}

func statusFinalEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.JetStreamTypes {
	// TODO: This method will be removed once CRD implementation is completed
	var finalEventTypes []eventingv1alpha2.JetStreamTypes
	for _, i := range typeInfos {
		finalEventTypes = append(finalEventTypes, eventingv1alpha2.JetStreamTypes{OriginalType: i.OriginalType, ConsumerName: i.ProcessedType})
	}
	return finalEventTypes
}

// setEmsSubscriptionStatus sets the status of bebSubscription in ev2Subscription.
func setEmsSubscriptionStatus(subscription *eventingv1alpha2.Subscription, eventMeshSubscription *types.Subscription) bool {
	var statusChanged = false
	if subscription.Status.Backend.EmsSubscriptionStatus == nil {
		subscription.Status.Backend.EmsSubscriptionStatus = &eventingv1alpha2.EmsSubscriptionStatus{}
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.Status != string(eventMeshSubscription.SubscriptionStatus) {
		subscription.Status.Backend.EmsSubscriptionStatus.Status = string(eventMeshSubscription.SubscriptionStatus)
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.StatusReason != eventMeshSubscription.SubscriptionStatusReason {
		subscription.Status.Backend.EmsSubscriptionStatus.StatusReason = eventMeshSubscription.SubscriptionStatusReason
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery != eventMeshSubscription.LastSuccessfulDelivery {
		subscription.Status.Backend.EmsSubscriptionStatus.LastSuccessfulDelivery = eventMeshSubscription.LastSuccessfulDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery != eventMeshSubscription.LastFailedDelivery {
		subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDelivery = eventMeshSubscription.LastFailedDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason != eventMeshSubscription.LastFailedDeliveryReason {
		subscription.Status.Backend.EmsSubscriptionStatus.LastFailedDeliveryReason = eventMeshSubscription.LastFailedDeliveryReason
		statusChanged = true
	}
	return statusChanged
}
