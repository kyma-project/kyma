package backend

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"reflect"
)

func isBackendStatusEqual(oldStatus, newStatus eventingv1alpha1.EventingBackendStatus) bool {
	oldStatusWithoutCond := oldStatus.DeepCopy()
	newStatusWithoutCond := newStatus.DeepCopy()

	// remove conditions, so that we don't compare them
	oldStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}
	newStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}

	return reflect.DeepEqual(oldStatusWithoutCond, newStatusWithoutCond) && eventingv1alpha1.ConditionsEquals(oldStatus.Conditions, newStatus.Conditions)
}
