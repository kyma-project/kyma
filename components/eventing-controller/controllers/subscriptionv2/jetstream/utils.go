package jetstream

import (
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// isInDeletion checks if the subscription needs to be deleted.
func isInDeletion(subscription *eventingv1alpha2.Subscription) bool {
	return !subscription.ObjectMeta.DeletionTimestamp.IsZero()
}

// containsFinalizer checks if the subscription contains our Finalizer.
func containsFinalizer(sub *eventingv1alpha2.Subscription) bool {
	return utils.ContainsString(sub.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer)
}
