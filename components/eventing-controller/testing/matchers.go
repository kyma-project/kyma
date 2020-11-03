package testing

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
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

func HaveValidAPIRule(s *eventingv1alpha1.Subscription) GomegaMatcher {
	return WithTransform(func(apiRule *apigatewayv1alpha1.APIRule) bool {
		hasOwnRef, hasRule := false, false
		for _, or := range apiRule.OwnerReferences {
			if or.Name == s.Name && or.UID == s.UID {
				hasOwnRef = true
				break
			}
		}
		if !hasOwnRef {
			return false
		}

		sURL, err := url.ParseRequestURI(s.Spec.Sink)
		if err != nil {
			log.Panic("failed to parse sub sink URI", err)
		}

		sURLSplitArr := strings.Split(sURL.Host, ".")
		if len(sURLSplitArr) != 5 {
			log.Panic("sink URL is invalid")
		}
		_, subscriberSvcName := sURLSplitArr[1], sURLSplitArr[0]
		if sURL.Path == "" {
			sURL.Path = "/"
		}
		for _, rule := range apiRule.Spec.Rules {
			if rule.Path == sURL.Path {
				if len(rule.Methods) != 2 {
					break
				}
				acceptableMethods := []string{http.MethodPost, http.MethodOptions}
				if !(containsString(acceptableMethods, rule.Methods[0]) && containsString(acceptableMethods, rule.Methods[1])) {
					break
				}
				if len(rule.AccessStrategies) != 1 {
					break
				}
				accessStrategy := rule.AccessStrategies[0]
				if accessStrategy.Handler.Name != "oauth2_introspection" {
					break
				}
				hasRule = true
				break
			}
		}
		if !hasRule {
			return false
		}

		if *apiRule.Spec.Gateway != constants.ClusterLocalAPIGateway {
			return false
		}

		expectedLabels := map[string]string{
			constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
			constants.ControllerServiceLabelKey:  subscriberSvcName,
		}
		if !reflect.DeepEqual(expectedLabels, apiRule.Labels) {
			return false
		}

		if subscriberSvcName != *apiRule.Spec.Service.Name {
			return false
		}
		svcPort, err := handlers.ConvertStringPortUInt32Port(*sURL)
		if err != nil {
			log.Panic("failed to convert sink port to uint32 port")
		}

		if svcPort != *apiRule.Spec.Service.Port {
			return false
		}

		return true
	}, BeTrue())
}

func containsString(list []string, contains string) bool {
	for _, elem := range list {
		if elem == contains {
			return true
		}
	}
	return false
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
