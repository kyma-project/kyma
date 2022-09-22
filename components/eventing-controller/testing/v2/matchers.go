package v2

import (
	"reflect"

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

//func HaveSubsConfiguration(subsConf *eventingv1alpha2.SubscriptionConfig) gomegatypes.GomegaMatcher {
//	return WithTransform(func(s *eventingv1alpha2.Subscription) *eventingv1alpha2.SubscriptionConfig {
//		return s.Status.Config
//	}, Equal(subsConf))
//}

func IsAnEmptySubscription() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		emptySub := eventingv1alpha2.Subscription{}
		return reflect.DeepEqual(*s, emptySub)
	}, BeTrue())
}

//
// Subscription matchers
//

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

func HaveSubscriptionNotReady() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) bool {
		return s.Status.Ready
	}, BeFalse())
}

func HaveCondition(condition eventingv1alpha2.Condition) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.Condition { return s.Status.Conditions }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Type":    Equal(condition.Type),
		"Reason":  Equal(condition.Reason),
		"Message": Equal(condition.Message),
		"Status":  Equal(condition.Status),
	})))
}

//func HaveConditionBadSubject() gomegatypes.GomegaMatcher {
//	condition := eventingv1alpha2.MakeCondition(
//		eventingv1alpha2.ConditionSubscriptionActive,
//		eventingv1alpha2.ConditionReasonNATSSubscriptionNotActive,
//		corev1.ConditionFalse, "failed to get clean subjects: "+nats.ErrBadSubject.Error(),
//	)
//	return HaveCondition(condition)
//}

//func HaveConditionInvalidPrefix() gomegatypes.GomegaMatcher {
//	condition := eventingv1alpha2.MakeCondition(
//		eventingv1alpha2.ConditionSubscriptionActive,
//		eventingv1alpha2.ConditionReasonNATSSubscriptionNotActive,
//		corev1.ConditionFalse, "failed to get clean subjects: prefix not found",
//	)
//	return HaveCondition(condition)
//}

func HaveCleanEventTypes(cleanEventTypes []eventingv1alpha2.EventType) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventType {
			return s.Status.Types
		},
		Equal(cleanEventTypes))
}

func HaveCleanEventTypesEmpty() gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha2.Subscription) []eventingv1alpha2.EventType {
			return s.Status.Types
		},
		BeEmpty())
}
