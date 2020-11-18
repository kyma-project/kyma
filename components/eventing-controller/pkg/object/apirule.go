package object

import (
	"net/url"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OAuthHandlerName OAuth handler name supported in Kyma
const OAuthHandlerName = "oauth2_introspection"

// NewAPIRule creates a APIRule object.
func NewAPIRule(ns, namePrefix string, opts ...Option) *apigatewayv1alpha1.APIRule {
	s := &apigatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    ns,
			GenerateName: namePrefix,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ApplyExistingAPIRuleAttributes copies some important attributes from a given
// source APIRule to a destination APIRule.
func ApplyExistingAPIRuleAttributes(src, dst *apigatewayv1alpha1.APIRule) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.Name = src.Name
	dst.GenerateName = ""
	dst.ResourceVersion = src.ResourceVersion
	dst.Spec.Service.Host = src.Spec.Service.Host
	// preserve status to avoid resetting conditions
	dst.Status = src.Status
}

// WithService sets the Service of an APIRule
func WithService(host, svcName string, port uint32) Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1alpha1.APIRule)
		isExternal := true
		apiService := apigatewayv1alpha1.Service{
			Name:       &svcName,
			Port:       &port,
			Host:       &host,
			IsExternal: &isExternal,
		}
		d.Spec.Service = &apiService
	}
}

// WithGateway sets the gateway of an APIRule
func WithGateway(gw string) Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1alpha1.APIRule)
		d.Spec.Gateway = &gw
	}
}

// WithOwnerReference sets the OwnerReferences of an APIRule
func WithOwnerReference(subs []eventingv1alpha1.Subscription) Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1alpha1.APIRule)
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

// WithRules sets the rules of an APIRule for all Subscriptions for a subscriber
func WithRules(subs []eventingv1alpha1.Subscription, methods ...string) Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1alpha1.APIRule)
		handler := oryv1alpha1.Handler{
			Name: OAuthHandlerName,
		}
		authenticator := &oryv1alpha1.Authenticator{
			Handler: &handler,
		}
		accessStrategies := []*oryv1alpha1.Authenticator{
			authenticator,
		}
		rules := make([]apigatewayv1alpha1.Rule, 0)
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
		uniquePaths := removeDuplicateValues(paths)
		for _, path := range uniquePaths {
			rule := apigatewayv1alpha1.Rule{
				Path:             path,
				Methods:          methods,
				AccessStrategies: accessStrategies,
			}
			rules = append(rules, rule)
		}
		d.Spec.Rules = rules
	}
}

func removeDuplicateValues(values []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0)

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range values {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// WithLabels sets the labels for an APIRule
func WithLabels(labels map[string]string) Option {
	return func(o metav1.Object) {
		d := o.(*apigatewayv1alpha1.APIRule)
		d.Labels = labels
	}
}
