package eventmesh

import (
	"fmt"
	"strings"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// getEventMeshSubject appends the prefix to subject.
func getEventMeshSubject(source, subject, eventMeshPrefix string) string {
	return fmt.Sprintf("%s.%s.%s", eventMeshPrefix, source, subject)
}

// isEventTypeSegmentsOverLimit checks if the number of segments in event type
// do not exceed eventTypeSegmentsLimit.
func isEventTypeSegmentsOverLimit(eventType string) bool {
	segments := strings.Split(eventType, ".")
	return len(segments) > eventTypeSegmentsLimit
}

func updateHashesInStatus(kymaSubscription *eventingv1alpha2.Subscription,
	eventMeshLocalSubscription *types.Subscription, eventMeshServerSubscription *types.Subscription) error {
	if err := setEventMeshLocalSubHashInStatus(kymaSubscription, eventMeshLocalSubscription); err != nil {
		return err
	}
	if err := setEventMeshServerSubHashInStatus(kymaSubscription, eventMeshServerSubscription); err != nil {
		return err
	}
	return setWebhookAuthHashInStatus(kymaSubscription, eventMeshLocalSubscription.WebhookAuth)
}

// setEventMeshLocalSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func setEventMeshLocalSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription,
	eventMeshSubscription *types.Subscription) error {
	// generate hash
	newHash, err := backendutils.GetHash(eventMeshSubscription)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.Ev2hash = newHash

	// set EventMesh Subscription hash without the webhook auth config
	cleanedEventMeshSub := backendutils.GetCleanedEventMeshSubscription(eventMeshSubscription)
	newHash, err = backendutils.GetHash(cleanedEventMeshSub)
	if err != nil {
		return err
	}
	kymaSubscription.Status.Backend.EventMeshLocalHash = newHash

	return nil
}

func setWebhookAuthHashInStatus(kymaSubscription *eventingv1alpha2.Subscription, webhookAuth *types.WebhookAuth) error {
	newHash, err := backendutils.GetWebhookAuthHash(webhookAuth)
	if err != nil {
		return err
	}
	kymaSubscription.Status.Backend.WebhookAuthHash = newHash
	return nil
}

// setEventMeshServerSubHashInStatus sets the hash for EventMesh local sub in Kyma Sub status.
func setEventMeshServerSubHashInStatus(kymaSubscription *eventingv1alpha2.Subscription,
	eventMeshSubscription *types.Subscription) error {
	// clean up the server sub object from extra info
	cleanedEventMeshSub := backendutils.GetCleanedEventMeshSubscription(eventMeshSubscription)
	// generate hash
	newHash, err := backendutils.GetHash(cleanedEventMeshSub)
	if err != nil {
		return err
	}

	// set hash in status
	kymaSubscription.Status.Backend.EventMeshHash = newHash
	return nil
}

func statusCleanEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.EventType {
	var cleanEventTypes []eventingv1alpha2.EventType
	for _, i := range typeInfos {
		cleanEventTypes = append(cleanEventTypes, eventingv1alpha2.EventType{OriginalType: i.OriginalType,
			CleanType: i.CleanType})
	}
	return cleanEventTypes
}

func statusFinalEventTypes(typeInfos []backendutils.EventTypeInfo) []eventingv1alpha2.EventMeshTypes {
	var finalEventTypes []eventingv1alpha2.EventMeshTypes
	for _, i := range typeInfos {
		finalEventTypes = append(finalEventTypes, eventingv1alpha2.EventMeshTypes{
			OriginalType:  i.OriginalType,
			EventMeshType: i.ProcessedType,
		})
	}
	return finalEventTypes
}

// setEmsSubscriptionStatus sets the status of EventMesh Subscription in ev2Subscription.
func setEmsSubscriptionStatus(subscription *eventingv1alpha2.Subscription,
	eventMeshSubscription *types.Subscription) bool {
	var statusChanged = false
	if subscription.Status.Backend.EventMeshSubscriptionStatus == nil {
		subscription.Status.Backend.EventMeshSubscriptionStatus = &eventingv1alpha2.EventMeshSubscriptionStatus{}
	}
	if subscription.Status.Backend.EventMeshSubscriptionStatus.Status != string(eventMeshSubscription.SubscriptionStatus) {
		subscription.Status.Backend.EventMeshSubscriptionStatus.Status = string(eventMeshSubscription.SubscriptionStatus)
		statusChanged = true
	}
	if subscription.Status.Backend.EventMeshSubscriptionStatus.StatusReason !=
		eventMeshSubscription.SubscriptionStatusReason {
		subscription.Status.Backend.EventMeshSubscriptionStatus.StatusReason =
			eventMeshSubscription.SubscriptionStatusReason
		statusChanged = true
	}
	if subscription.Status.Backend.EventMeshSubscriptionStatus.LastSuccessfulDelivery !=
		eventMeshSubscription.LastSuccessfulDelivery {
		subscription.Status.Backend.EventMeshSubscriptionStatus.LastSuccessfulDelivery =
			eventMeshSubscription.LastSuccessfulDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDelivery !=
		eventMeshSubscription.LastFailedDelivery {
		subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDelivery =
			eventMeshSubscription.LastFailedDelivery
		statusChanged = true
	}
	if subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDeliveryReason !=
		eventMeshSubscription.LastFailedDeliveryReason {
		subscription.Status.Backend.EventMeshSubscriptionStatus.LastFailedDeliveryReason =
			eventMeshSubscription.LastFailedDeliveryReason
		statusChanged = true
	}
	return statusChanged
}
