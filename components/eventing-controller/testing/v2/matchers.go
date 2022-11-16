package v2

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"strconv"

	. "github.com/onsi/gomega"         // nolint
	. "github.com/onsi/gomega/gstruct" // nolint
	gomegatypes "github.com/onsi/gomega/types"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

//
// string matchers
//

func HaveSubscriptionName(name string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) string { return s.Name }, Equal(name))
}

func HaveSubscriptionSink(sink string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) string { return s.Spec.Sink }, Equal(sink))
}

func HaveSubscriptionFinalizer(finalizer string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) []string { return s.ObjectMeta.Finalizers }, ContainElement(finalizer))
}

func HaveSubscriptionLabels(labels map[string]string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) map[string]string { return s.Labels }, Equal(labels))
}

func IsAnEmptySubscription() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		emptySub := eventingv1alpha2.Subscription{}
		return reflect.DeepEqual(*s, emptySub)
	}, BeTrue())
}

func HaveNoneEmptyAPIRuleName() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) string {
		return s.Status.Backend.APIRuleName
	}, Not(BeEmpty()))
}

func HaveAPIRuleName(name string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		return s.Status.Backend.APIRuleName == name
	}, BeTrue())
}

func HaveSubscriptionReady() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		return s.Status.Ready
	}, BeTrue())
}
func HaveTypes(types []string) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []string {
			return s.Spec.Types
		},
		Equal(types))
}

func HaveMaxInFlight(maxInFlight int) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		return s.Spec.Config[eventingv1alpha2.MaxInFlightMessages] == strconv.Itoa(maxInFlight)
	}, BeTrue())
}

func HaveSubscriptionNotReady() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		return s.Status.Ready
	}, BeFalse())
}

func HaveCondition(condition eventingv1alpha2.Condition) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.Condition { return s.Status.Conditions },
		ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
			"Type":    Equal(condition.Type),
			"Reason":  Equal(condition.Reason),
			"Message": Equal(condition.Message),
			"Status":  Equal(condition.Status),
		})))
}

func HaveSubscriptionActiveCondition() gomegatypes.GomegaMatcher {
	return HaveCondition(eventingv1alpha2.MakeCondition(
		eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonSubscriptionActive,
		corev1.ConditionTrue, ""))
}

func HaveAPIRuleTrueStatusCondition() gomegatypes.GomegaMatcher {
	return HaveCondition(eventingv1alpha2.MakeCondition(
		eventingv1alpha2.ConditionAPIRuleStatus,
		eventingv1alpha2.ConditionReasonAPIRuleStatusReady,
		corev1.ConditionTrue,
		"",
	))
}

func HaveCleanEventTypes(cleanEventTypes []eventingv1alpha2.EventType) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventType {
			return s.Status.Types
		},
		Equal(cleanEventTypes))
}

func DefaultReadyCondition() eventingv1alpha2.Condition {
	return eventingv1alpha2.MakeCondition(
		eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")
}

func HaveStatusTypes(cleanEventTypes []eventingv1alpha2.EventType) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventType {
			return s.Status.Types
		},
		Equal(cleanEventTypes))
}

func HaveNotFoundSubscription() gomegatypes.GomegaMatcher {
	return WithTransform(func(isDeleted bool) bool { return isDeleted }, BeTrue())
}

func HaveEvent(event corev1.Event) gomegatypes.GomegaMatcher {
	return WithTransform(func(l corev1.EventList) []corev1.Event { return l.Items }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Reason":  Equal(event.Reason),
		"Message": Equal(event.Message),
		"Type":    Equal(event.Type),
	})))
}

func HaveCleanEventTypesEmpty() gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventType {
			return s.Status.Types
		},
		BeEmpty())
}

func HaveEventMeshTypes(emsTypes []eventingv1alpha2.EventMeshTypes) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventMeshTypes {
			return s.Status.Backend.EmsTypes
		},
		Equal(emsTypes))
}
