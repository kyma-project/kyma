package handlers

import (
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

type APIRuleOption func(rule *apigatewayv1alpha1.APIRule)

// NewAPIRule returns a valid APIRule
func NewAPIRule(opts ...APIRuleOption) *apigatewayv1alpha1.APIRule {
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

func WithService(apiRule *apigatewayv1alpha1.APIRule) {
	port := uint32(9999)
	isExternal := true
	host := "foo-host"
	svcName := "foo-svc"
	apiRule.Spec.Service = &apigatewayv1alpha1.Service{
		Name:       &svcName,
		Port:       &port,
		Host:       &host,
		IsExternal: &isExternal,
	}
}

func WithGateway(apiRule *apigatewayv1alpha1.APIRule) {
	gateway := "foo.gateway"
	apiRule.Spec.Gateway = &gateway
}

func WithPath(apiRule *apigatewayv1alpha1.APIRule) {
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

func WithoutPath(apiRule *apigatewayv1alpha1.APIRule) {
	handlerOAuth := "oauth2_introspection"
	handler := oryv1alpha1.Handler{
		Name: handlerOAuth,
	}
	authenticator := &oryv1alpha1.Authenticator{
		Handler: &handler,
	}
	apiRule.Spec.Rules = []apigatewayv1alpha1.Rule{
		{
			Path: "/",
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

func WithStatusReady(apiRule *apigatewayv1alpha1.APIRule) {
	statusOK := &apigatewayv1alpha1.APIRuleResourceStatus{
		Code:        apigatewayv1alpha1.StatusOK,
		Description: "",
	}

	apiRule.Status = apigatewayv1alpha1.APIRuleStatus{
		APIRuleStatus:        statusOK,
		VirtualServiceStatus: statusOK,
		AccessRuleStatus:     statusOK,
	}
}
