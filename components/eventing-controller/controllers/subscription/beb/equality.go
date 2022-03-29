package beb

import (
	"reflect"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func isSubscriptionStatusEqual(oldStatus, newStatus eventingv1alpha1.SubscriptionStatus) bool {
	oldStatusWithoutCond := oldStatus.DeepCopy()
	newStatusWithoutCond := newStatus.DeepCopy()

	// remove conditions, so that we don't compare them
	oldStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}
	newStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}

	return reflect.DeepEqual(oldStatusWithoutCond, newStatusWithoutCond) && eventingv1alpha1.ConditionsEquals(oldStatus.Conditions, newStatus.Conditions)
}
