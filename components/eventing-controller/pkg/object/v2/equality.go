package v2

import (
	"reflect"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

func IsSubscriptionStatusEqual(oldStatus, newStatus eventingv1alpha2.SubscriptionStatus) bool {
	oldStatusWithoutCond := oldStatus.DeepCopy()
	newStatusWithoutCond := newStatus.DeepCopy()

	// remove conditions, so that we don't compare them
	oldStatusWithoutCond.Conditions = []eventingv1alpha2.Condition{}
	newStatusWithoutCond.Conditions = []eventingv1alpha2.Condition{}

	return reflect.DeepEqual(oldStatusWithoutCond, newStatusWithoutCond) &&
		eventingv1alpha2.ConditionsEquals(oldStatus.Conditions, newStatus.Conditions)
}
