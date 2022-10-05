package v2

import (
	"net/url"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WithOwnerReference sets the OwnerReferences of an APIRule.
func WithOwnerReference(subs []eventingv1alpha2.Subscription) object.Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1beta1.APIRule)
		ownerRefs := make([]metav1.OwnerReference, 0)
		if len(subs) > 0 {
			for _, sub := range subs {
				blockOwnerDeletion := true
				ownerRef := metav1.OwnerReference{
					APIVersion:         sub.APIVersion,
					Kind:               sub.Kind,
					Name:               sub.Name,
					UID:                sub.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				}
				ownerRefs = append(ownerRefs, ownerRef)
			}
		}

		d.OwnerReferences = ownerRefs
	}
}

// WithRules sets the rules of an APIRule for all Subscriptions for a subscriber.
func WithRules(subs []eventingv1alpha2.Subscription, svc apigatewayv1beta1.Service, methods ...string) object.Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1beta1.APIRule)
		handler := apigatewayv1beta1.Handler{
			Name: object.OAuthHandlerName,
		}
		authenticator := &apigatewayv1beta1.Authenticator{
			Handler: &handler,
		}
		accessStrategies := []*apigatewayv1beta1.Authenticator{
			authenticator,
		}
		rules := make([]apigatewayv1beta1.Rule, 0)
		paths := make([]string, 0)
		for _, sub := range subs {
			hostURL, err := url.ParseRequestURI(sub.Spec.Sink)
			if err != nil {
				// It's ok as the relevant subscription will have a valid cluster local URL in the same namespace
				continue
			}
			if hostURL.Path == "" {
				paths = append(paths, "/")
			} else {
				paths = append(paths, hostURL.Path)
			}
		}
		uniquePaths := object.RemoveDuplicateValues(paths)
		for _, path := range uniquePaths {
			rule := apigatewayv1beta1.Rule{
				Path:             path,
				Methods:          methods,
				AccessStrategies: accessStrategies,
				Service:          &svc,
			}
			rules = append(rules, rule)
		}
		d.Spec.Rules = rules
	}
}
