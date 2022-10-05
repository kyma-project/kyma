package jetstream

import (
	"strings"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	corev1 "k8s.io/api/core/v1"
)

// isInDeletion checks if the subscription needs to be deleted.
func isInDeletion(subscription *eventingv1alpha2.Subscription) bool {
	return !subscription.ObjectMeta.DeletionTimestamp.IsZero()
}

// containsFinalizer checks if the subscription contains our Finalizer.
func containsFinalizer(sub *eventingv1alpha2.Subscription) bool {
	return utils.ContainsString(sub.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer)
}

// missingSubscriptionErr checks if the error reports about missing NATS subscription in js.subscriptions map.
func missingSubscriptionErr(err error) bool {
	return strings.Contains(err.Error(), jetstreamv2.MissingNATSSubscriptionMsg)
}

// setSubReadyStatus returns true if the subscription ready status has changed.
func setSubReadyStatus(desiredSubscriptionStatus *eventingv1alpha2.SubscriptionStatus, isReady bool) bool {
	if desiredSubscriptionStatus.Ready != isReady {
		desiredSubscriptionStatus.Ready = isReady
		return true
	}
	return false
}

//----------------------------------------
// Condition utils
//----------------------------------------

// initializeDesiredConditions initializes the required conditions for the subscription status.
func initializeDesiredConditions() []eventingv1alpha2.Condition {
	desiredConditions := make([]eventingv1alpha2.Condition, 0)
	condition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonNATSSubscriptionNotActive, corev1.ConditionFalse, "")
	desiredConditions = append(desiredConditions, condition)
	return desiredConditions
}

// setConditionSubscriptionActive updates the ConditionSubscriptionActive condition if the error is nil.
func setConditionSubscriptionActive(desiredConditions []eventingv1alpha2.Condition, error error) {
	for key, c := range desiredConditions {
		if c.Type == eventingv1alpha2.ConditionSubscriptionActive {
			if error == nil {
				desiredConditions[key].Status = corev1.ConditionTrue
				desiredConditions[key].Reason = eventingv1alpha2.ConditionReasonNATSSubscriptionActive
			} else {
				desiredConditions[key].Message = error.Error()
			}
		}
	}
}
