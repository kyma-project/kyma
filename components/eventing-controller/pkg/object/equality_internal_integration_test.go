//go:build integration
// +build integration

package object

import (
	"net/http"
	"testing"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiRuleEqual(t *testing.T) {
	svc := "svc"
	port := uint32(9999)
	host := "host"
	isExternal := true
	gateway := "foo.gateway"
	labels := map[string]string{
		"foo": "bar",
	}
	handler := &apigatewayv1beta1.Handler{
		Name: "handler",
	}
	rule := apigatewayv1beta1.Rule{
		Path: "path",
		Methods: []string{
			http.MethodPost,
		},
		AccessStrategies: []*apigatewayv1beta1.Authenticator{
			{
				Handler: handler,
			},
		},
	}
	apiRule := apigatewayv1beta1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels:    labels,
		},
		Spec: apigatewayv1beta1.APIRuleSpec{
			Service: &apigatewayv1beta1.Service{
				Name:       &svc,
				Port:       &port,
				IsExternal: &isExternal,
			},
			Host:    &host,
			Gateway: &gateway,
			Rules:   []apigatewayv1beta1.Rule{rule},
		},
	}

	testCases := map[string]struct {
		prep   func() *apigatewayv1beta1.APIRule
		expect bool
	}{
		"should be equal when svc, gateway, owner ref, rules are same": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				return apiRuleCopy
			},
			expect: true,
		},
		"should be unequal when svc name is diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newSvcName := "new"
				apiRuleCopy.Spec.Service.Name = &newSvcName
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when svc port is diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newSvcPort := uint32(8080)
				apiRuleCopy.Spec.Service.Port = &newSvcPort
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when isExternal is diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newIsExternal := false
				apiRuleCopy.Spec.Service.IsExternal = &newIsExternal
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when gateway is diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newGateway := "new-gw"
				apiRuleCopy.Spec.Gateway = &newGateway
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when labels are diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newLabels := map[string]string{
					"new-foo": "new-bar",
				}
				apiRuleCopy.Labels = newLabels
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when path is diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newRule.Path = "new-path"
				apiRuleCopy.Spec.Rules = []apigatewayv1beta1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when methods are diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newRule.Methods = []string{http.MethodOptions}
				apiRuleCopy.Spec.Rules = []apigatewayv1beta1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when handlers are diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newHandler := &apigatewayv1beta1.Handler{
					Name: "foo",
				}
				newRule.AccessStrategies = []*apigatewayv1beta1.Authenticator{
					{
						Handler: newHandler,
					},
				}
				apiRuleCopy.Spec.Rules = []apigatewayv1beta1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when OwnerReferences are diff": {
			prep: func() *apigatewayv1beta1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newOwnerRef := metav1.OwnerReference{
					APIVersion: "foo",
					Kind:       "foo",
					Name:       "foo",
					UID:        "uid",
				}
				apiRuleCopy.OwnerReferences = []metav1.OwnerReference{
					newOwnerRef,
				}
				return apiRuleCopy
			},
			expect: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testAPIRule := tc.prep()
			if apiRuleEqual(&apiRule, testAPIRule) != tc.expect {
				t.Errorf("expected output to be %t", tc.expect)
			}
		})
	}
}
