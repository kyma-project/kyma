package testing

import (
	"reflect"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	. "github.com/onsi/gomega"         // nolint
	. "github.com/onsi/gomega/gstruct" // nolint
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
)

//
// string matchers
//

func HaveSubscriptionName(name string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) string { return s.Name }, Equal(name))
}

func HaveSubscriptionSink(sink string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) string { return s.Spec.Sink }, Equal(sink))
}

func HaveSubscriptionFinalizer(finalizer string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) []string { return s.ObjectMeta.Finalizers }, ContainElement(finalizer))
}

func HaveSubscriptionLabels(labels map[string]string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) map[string]string { return s.Labels }, Equal(labels))
}

func HaveNotFoundSubscription(isReallyDeleted bool) gomegatypes.GomegaMatcher {
	return WithTransform(func(isDeleted bool) bool { return isDeleted }, Equal(isReallyDeleted))
}

func HaveSubsConfiguration(subsConf *eventingv1alpha1.SubscriptionConfig) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) *eventingv1alpha1.SubscriptionConfig {
		return s.Status.Config
	}, Equal(subsConf))
}

func IsAnEmptySubscription() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) bool {
		emptySub := eventingv1alpha1.Subscription{}
		return reflect.DeepEqual(*s, emptySub)
	}, BeTrue())
}

//
// APIRule matchers
//

func HaveNotEmptyAPIRule() gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) types.UID {
		return a.UID
	}, Not(BeEmpty()))
}

func HaveNotEmptyHost() gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) bool {
		return a.Spec.Service != nil && a.Spec.Service.Host != nil
	}, BeTrue())
}

func HaveAPIRuleGateway(gateway string) gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) string {
		if a.Spec.Gateway == nil {
			return ""
		}
		return *a.Spec.Gateway
	}, Equal(gateway))
}

func HaveAPIRuleLabels(labels map[string]string) gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) map[string]string {
		return a.Labels
	}, Equal(labels))
}

func HaveAPIRuleService(serviceName string, port uint32, domain string) gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) apigatewayv1alpha1.Service {
		if a.Spec.Service == nil {
			return apigatewayv1alpha1.Service{}
		}
		return *a.Spec.Service
	}, MatchFields(IgnoreMissing|IgnoreExtras, Fields{
		"Port":       PointTo(Equal(port)),
		"Name":       PointTo(Equal(serviceName)),
		"Host":       PointTo(ContainSubstring(domain)),
		"IsExternal": PointTo(BeTrue()),
	}),
	)
}

func HaveAPIRuleSpecRules(ruleMethods []string, accessStrategy, path string) gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) []apigatewayv1alpha1.Rule {
		return a.Spec.Rules
	}, ContainElement(
		MatchFields(IgnoreExtras|IgnoreMissing, Fields{
			"Methods":          ConsistOf(ruleMethods),
			"AccessStrategies": ConsistOf(haveAPIRuleAccessStrategies(accessStrategy)),
			"Gateway":          Equal(constants.ClusterLocalAPIGateway),
			"Path":             Equal(path),
		}),
	))
}

func haveAPIRuleAccessStrategies(accessStrategy string) gomegatypes.GomegaMatcher {
	return WithTransform(func(a *apigatewayv1alpha1.Authenticator) string {
		return a.Name
	}, Equal(accessStrategy))
}

func HaveAPIRuleOwnersRefs(uids ...types.UID) gomegatypes.GomegaMatcher {
	return WithTransform(func(a apigatewayv1alpha1.APIRule) []types.UID {
		ownerRefUIDs := make([]types.UID, 0, len(a.OwnerReferences))
		for _, ownerRef := range a.OwnerReferences {
			ownerRefUIDs = append(ownerRefUIDs, ownerRef.UID)
		}
		return ownerRefUIDs
	}, Equal(uids))
}

//
// Subscription matchers
//

func HaveNoneEmptyAPIRuleName() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) string {
		return s.Status.APIRuleName
	}, Not(BeEmpty()))
}

func HaveAPIRuleName(name string) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) bool {
		return s.Status.APIRuleName == name
	}, BeTrue())
}

func HaveSubscriptionReady() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) bool {
		return s.Status.Ready
	}, BeTrue())
}

func HaveSubscriptionNotReady() gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) bool {
		return s.Status.Ready
	}, BeFalse())
}

func HaveCondition(condition eventingv1alpha1.Condition) gomegatypes.GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) []eventingv1alpha1.Condition { return s.Status.Conditions }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Type":    Equal(condition.Type),
		"Reason":  Equal(condition.Reason),
		"Message": Equal(condition.Message),
		"Status":  Equal(condition.Status),
	})))
}

func HaveCleanEventTypes(cleanEventTypes []string) gomegatypes.GomegaMatcher {
	return WithTransform(
		func(s *eventingv1alpha1.Subscription) []string {
			return s.Status.CleanEventTypes
		},
		Equal(cleanEventTypes))
}

func HaveEvent(event corev1.Event) gomegatypes.GomegaMatcher {
	return WithTransform(func(l corev1.EventList) []corev1.Event { return l.Items }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Reason":  Equal(event.Reason),
		"Message": Equal(event.Message),
		"Type":    Equal(event.Type),
	})))
}

func IsK8sUnprocessableEntity() gomegatypes.GomegaMatcher {
	return WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, Equal(metav1.StatusReasonInvalid))
}

//
// int matchers
//

func BeGreaterThanOrEqual(a int) gomegatypes.GomegaMatcher {
	return WithTransform(func(b int) bool { return b >= a }, BeTrue())
}

func HaveValidClientID(clientIDKey, clientID string) gomegatypes.GomegaMatcher {
	return WithTransform(func(secret *corev1.Secret) bool {
		if secret != nil {
			return string(secret.Data[clientIDKey]) == clientID
		}
		return false
	}, BeTrue())
}

func HaveValidClientSecret(clientSecretKey, clientSecret string) gomegatypes.GomegaMatcher {
	return WithTransform(func(secret *corev1.Secret) bool {
		if secret != nil {
			return string(secret.Data[clientSecretKey]) == clientSecret
		}
		return false
	}, BeTrue())
}

func HaveValidTokenEndpoint(tokenEndpointKey, tokenEndpoint string) gomegatypes.GomegaMatcher {
	return WithTransform(func(secret *corev1.Secret) bool {
		if secret != nil {
			return string(secret.Data[tokenEndpointKey]) == tokenEndpoint
		}
		return false
	}, BeTrue())
}

func HaveValidEMSPublishURL(emsPublishURLKey, emsPublishURL string) gomegatypes.GomegaMatcher {
	return WithTransform(func(secret *corev1.Secret) bool {
		if secret != nil {
			return string(secret.Data[emsPublishURLKey]) == emsPublishURL
		}
		return false
	}, BeTrue())
}

func HaveValidBEBNamespace(bebNamespaceKey, namespace string) gomegatypes.GomegaMatcher {
	return WithTransform(func(secret *corev1.Secret) bool {
		if secret != nil {
			return string(secret.Data[bebNamespaceKey]) == namespace
		}
		return false
	}, BeTrue())
}
