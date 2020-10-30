package handlers

import (
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

type APIRuleOption func(rule *apigatewayv1alpha1.APIRule)

// newAPIRule returns a valid APIRule
func newAPIRule(opts ...APIRuleOption) *apigatewayv1alpha1.APIRule {
	apiRule := &apigatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}

	for _, opt := range opts {
		opt(apiRule)
	}
	return apiRule
}

func withService(apiRule *apigatewayv1alpha1.APIRule) {
	port := uint32(9999)
	isExternal := true
	host := "foo-host"
	apiRule.Spec.Service = &apigatewayv1alpha1.Service{
		Port:       &port,
		Host:       &host,
		IsExternal: &isExternal,
	}
}

func withGateway(apiRule *apigatewayv1alpha1.APIRule) {
	gateway := "foo-gateway"
	apiRule.Spec.Gateway = &gateway
}

func withPath(apiRule *apigatewayv1alpha1.APIRule) {
	handlerOAuth := "oauth2_introspection"
	handler := oryv1alpha1.Handler{
		Name: handlerOAuth,
	}
	authenticator := &oryv1alpha1.Authenticator{
		Handler: &handler,
	}
	apiRule.Spec.Rules = []apigatewayv1alpha1.Rule{
		{
			Path: "/path",
			Methods: []string{
				http.MethodPost,
				http.MethodOptions,
			},
			AccessStrategies: []*oryv1alpha1.Authenticator{
				authenticator,
			},
		},
	}
}
