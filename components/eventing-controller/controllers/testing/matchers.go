package testing

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func HaveSubscriptionName(name string) GomegaMatcher {
	return WithTransform(func(s eventingv1alpha1.Subscription) string { return s.Name }, Equal(name))
}

func HaveSubscriptionFinalizer(finalizer string) GomegaMatcher {
	return WithTransform(func(s eventingv1alpha1.Subscription) []string { return s.ObjectMeta.Finalizers }, ContainElement(finalizer))
}

func HaveSubscriptionReady() GomegaMatcher {
	return WithTransform(func(s eventingv1alpha1.Subscription) bool {
		return s.Status.Ready
	}, BeTrue())
}

func HaveCondition(condition eventingv1alpha1.Condition) GomegaMatcher {
	return WithTransform(func(s eventingv1alpha1.Subscription) []eventingv1alpha1.Condition { return s.Status.Conditions }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Type":    Equal(condition.Type),
		"Reason":  Equal(condition.Reason),
		"Message": Equal(condition.Message),
	})))
}

func HaveEvent(event v1.Event) GomegaMatcher {
	return WithTransform(func(l v1.EventList) []v1.Event { return l.Items }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Reason":  Equal(event.Reason),
		"Message": Equal(event.Message),
		"Type":    Equal(event.Type),
	})))
}

func IsK8sUnprocessableEntity() GomegaMatcher {
	//  <*errors.StatusError | 0xc0001330e0>: {
	//     ErrStatus: {
	//         TypeMeta: {Kind: "", APIVersion: ""},
	//         ListMeta: {
	//             SelfLink: "",
	//             ResourceVersion: "",
	//             Continue: "",
	//             RemainingItemCount: nil,
	//         },
	//         Status: "Failure",
	//         Message: "Subscription.eventing.kyma-project.io \"test-valid-subscription-1\" is invalid: spec.filter: Invalid value: \"null\": spec.filter in body must be of type object: \"null\"",
	//         Reason: "Invalid",
	//         Details: {
	//             Name: "test-valid-subscription-1",
	//             Group: "eventing.kyma-project.io",
	//             Kind: "Subscription",
	//             UID: "",
	//             Causes: [
	//                 {
	//                     Type: "FieldValueInvalid",
	//                     Message: "Invalid value: \"null\": spec.filter in body must be of type object: \"null\"",
	//                     Field: "spec.filter",
	//                 },
	//             ],
	//             RetryAfterSeconds: 0,
	//         },
	//         Code: 422,
	//     },
	// }
	return WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, Equal(metav1.StatusReasonInvalid))
}
