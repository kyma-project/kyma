package subscription

import (
	"reflect"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// ConditionsEquals checks if two list of conditions are equal.
func ConditionsEquals(existing, expected []eventingv1alpha1.Condition) bool {
	// not equal if length is different
	if len(existing) != len(expected) {
		return false
	}

	// compile map of Conditions per ConditionType
	existingMap := make(map[eventingv1alpha1.ConditionType]eventingv1alpha1.Condition, len(existing))
	for _, value := range existing {
		existingMap[value.Type] = value
	}

	for _, value := range expected {
		if !ConditionEquals(existingMap[value.Type], value) {
			return false
		}
	}

	return true
}

// ConditionsEquals checks if two conditions are equal.
func ConditionEquals(existing, expected eventingv1alpha1.Condition) bool {
	isTypeEqual := existing.Type == expected.Type
	isStatusEqual := existing.Status == expected.Status
	isReasonEqual := existing.Reason == expected.Reason
	isMessageEqual := existing.Message == expected.Message

	if !isStatusEqual || !isReasonEqual || !isMessageEqual || !isTypeEqual {
		return false
	}

	return true
}

func IsSubscriptionStatusEqual(oldStatus, newStatus eventingv1alpha1.SubscriptionStatus) bool {
	oldStatusWithoutCond := oldStatus.DeepCopy()
	newStatusWithoutCond := newStatus.DeepCopy()

	// remove conditions, so that we don't compare them
	oldStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}
	newStatusWithoutCond.Conditions = []eventingv1alpha1.Condition{}

	return reflect.DeepEqual(oldStatusWithoutCond, newStatusWithoutCond) && ConditionsEquals(oldStatus.Conditions, newStatus.Conditions)
}
