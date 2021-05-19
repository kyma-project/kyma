package object

import (
	"net/http"
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	appsv1 "k8s.io/api/apps/v1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
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
	handler := &rulev1alpha1.Handler{
		Name: "handler",
	}
	rule := apigatewayv1alpha1.Rule{
		Path: "path",
		Methods: []string{
			http.MethodPost,
		},
		AccessStrategies: []*rulev1alpha1.Authenticator{
			{
				Handler: handler,
			},
		},
	}
	apiRule := apigatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels:    labels,
		},
		Spec: apigatewayv1alpha1.APIRuleSpec{
			Service: &apigatewayv1alpha1.Service{
				Name:       &svc,
				Port:       &port,
				Host:       &host,
				IsExternal: &isExternal,
			},
			Gateway: &gateway,
			Rules:   []apigatewayv1alpha1.Rule{rule},
		},
	}
	testCases := map[string]struct {
		prep   func() *apigatewayv1alpha1.APIRule
		expect bool
	}{
		"should be equal when svc, gateway, owner ref, rules are same": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				return apiRuleCopy
			},
			expect: true,
		},
		"should be unequal when svc name is diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newSvcName := "new"
				apiRuleCopy.Spec.Service.Name = &newSvcName
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when svc port is diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newSvcPort := uint32(8080)
				apiRuleCopy.Spec.Service.Port = &newSvcPort
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when isExternal is diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newIsExternal := false
				apiRuleCopy.Spec.Service.IsExternal = &newIsExternal
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when gateway is diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newGateway := "new-gw"
				apiRuleCopy.Spec.Gateway = &newGateway
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when labels are diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
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
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newRule.Path = "new-path"
				apiRuleCopy.Spec.Rules = []apigatewayv1alpha1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when methods are diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newRule.Methods = []string{http.MethodOptions}
				apiRuleCopy.Spec.Rules = []apigatewayv1alpha1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when handlers are diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
				apiRuleCopy := apiRule.DeepCopy()
				newRule := rule.DeepCopy()
				newHandler := &rulev1alpha1.Handler{
					Name: "foo",
				}
				newRule.AccessStrategies = []*rulev1alpha1.Authenticator{
					{
						Handler: newHandler,
					},
				}
				apiRuleCopy.Spec.Rules = []apigatewayv1alpha1.Rule{*newRule}
				return apiRuleCopy
			},
			expect: false,
		},
		"should be unequal when OwnerReferences are diff": {
			prep: func() *apigatewayv1alpha1.APIRule {
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
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestEventingBackendEqual(t *testing.T) {
	emptyBackend := eventingv1alpha1.EventingBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: eventingv1alpha1.EventingBackendSpec{},
	}

	testCases := map[string]struct {
		getBackend1    func() *eventingv1alpha1.EventingBackend
		getBackend2    func() *eventingv1alpha1.EventingBackend
		expectedResult bool
	}{
		"should be unequal if labels are different": {
			getBackend1: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			getBackend2: func() *eventingv1alpha1.EventingBackend {
				return emptyBackend.DeepCopy()
			},
			expectedResult: false,
		},
		"should be equal if labels are the same": {
			getBackend1: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			getBackend2: func() *eventingv1alpha1.EventingBackend {
				b := emptyBackend.DeepCopy()
				b.Name = "bar"
				b.Labels = map[string]string{"k1": "v1"}
				return b
			},
			expectedResult: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if eventingBackendEqual(tc.getBackend1(), tc.getBackend2()) != tc.expectedResult {
				t.Errorf("Expected output to be %t", tc.expectedResult)
			}
		})
	}
}

func TestEventingBackendStatusEqual(t *testing.T) {
	testCases := map[string]struct {
		backendStatus1 eventingv1alpha1.EventingBackendStatus
		backendStatus2 eventingv1alpha1.EventingBackendStatus
		expectedResult bool
	}{
		"should be unequal if ready status is different": {
			backendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady:               utils.BoolPtr(false),
				SubscriptionControllerReady: utils.BoolPtr(true),
				PublisherProxyReady:         utils.BoolPtr(true),
			},
			backendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady:               utils.BoolPtr(true),
				SubscriptionControllerReady: utils.BoolPtr(true),
				PublisherProxyReady:         utils.BoolPtr(true),
			},
			expectedResult: false,
		},
		"should be unequal if missing secret": {
			backendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady:               utils.BoolPtr(false),
				SubscriptionControllerReady: utils.BoolPtr(true),
				PublisherProxyReady:         utils.BoolPtr(true),
				BebSecretName:               "secret",
				BebSecretNamespace:          "default",
			},
			backendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady:               utils.BoolPtr(false),
				SubscriptionControllerReady: utils.BoolPtr(true),
				PublisherProxyReady:         utils.BoolPtr(true),
			},
			expectedResult: false,
		},
		"should be unequal if missing backend": {
			backendStatus1: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.NatsBackendType,
			},
			backendStatus2: eventingv1alpha1.EventingBackendStatus{},
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if eventingBackendStatusEqual(&tc.backendStatus1, &tc.backendStatus2) != tc.expectedResult {
				t.Errorf("Expected output to be %t", tc.expectedResult)
			}
		})
	}
}

func TestPublisherProxyDeploymentEqual(t *testing.T) {
	defaultNATSPublisher := deployment.NewNATSPublisherDeployment("publisher", "publisher", 1)
	defaultBEBPublisher := deployment.NewBEBPublisherDeployment("publisher", "publisher", 1)

	testCases := map[string]struct {
		getPublisher1  func() *appsv1.Deployment
		getPublisher2  func() *appsv1.Deployment
		expectedResult bool
	}{
		"should be equal if same default NATS publisher": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Name = "publisher1"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Name = "publisher2"
				return p
			},
			expectedResult: true,
		},
		"should be equal if same default BEB publisher": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultBEBPublisher.DeepCopy()
				p.Name = "publisher1"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultBEBPublisher.DeepCopy()
				p.Name = "publisher2"
				return p
			},
			expectedResult: true,
		},
		"should be unequal if publisher types are different": {
			getPublisher1: func() *appsv1.Deployment {
				return defaultBEBPublisher.DeepCopy()
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be unequal if publisher image changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Spec.Containers[0].Image = "new-publisher-img"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be unequal if env var changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Spec.Containers[0].Env[0].Value = "new-value"
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if publisherProxyDeploymentEqual(tc.getPublisher1(), tc.getPublisher2()) != tc.expectedResult {
				t.Errorf("Expected output to be %t", tc.expectedResult)
			}
		})
	}
}
