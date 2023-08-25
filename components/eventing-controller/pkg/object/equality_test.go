package object

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	apigatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
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
				t.Errorf("expected output to be %t", tc.expectedResult)
			}
		})
	}
}

func TestEventingBackendStatusEqual(t *testing.T) {
	testCases := []struct {
		name                string
		givenBackendStatus1 eventingv1alpha1.EventingBackendStatus
		givenBackendStatus2 eventingv1alpha1.EventingBackendStatus
		wantResult          bool
	}{
		{
			name: "should be unequal if ready status is different",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady: utils.BoolPtr(false),
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady: utils.BoolPtr(true),
			},
			wantResult: false,
		},
		{
			name: "should be unequal if missing secret",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady:      utils.BoolPtr(false),
				BEBSecretName:      "secret",
				BEBSecretNamespace: "default",
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady: utils.BoolPtr(false),
			},
			wantResult: false,
		},
		{
			name: "should be unequal if different secretName",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady:      utils.BoolPtr(false),
				BEBSecretName:      "secret",
				BEBSecretNamespace: "default",
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady:      utils.BoolPtr(false),
				BEBSecretName:      "secretnew",
				BEBSecretNamespace: "default",
			},
			wantResult: false,
		},
		{
			name: "should be unequal if different secretNamespace",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				EventingReady:      utils.BoolPtr(false),
				BEBSecretName:      "secret",
				BEBSecretNamespace: "default",
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				EventingReady:      utils.BoolPtr(false),
				BEBSecretName:      "secret",
				BEBSecretNamespace: "kyma-system",
			},
			wantResult: false,
		},
		{
			name: "should be unequal if missing backend",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.NatsBackendType,
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{},
			wantResult:          false,
		},
		{
			name: "should be unequal if different backend",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.NatsBackendType,
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.BEBBackendType,
			},
			wantResult: false,
		},
		{
			name: "should be unequal if conditions different",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionTrue},
				},
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionFalse},
				},
			},
			wantResult: false,
		},
		{
			name: "should be unequal if conditions missing",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionTrue},
				},
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{},
			},
			wantResult: false,
		},
		{
			name: "should be unequal if conditions different",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionTrue},
				},
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionControllerReady, Status: corev1.ConditionTrue},
				},
			},
			wantResult: false,
		},
		{
			name: "should be equal if the status are the same",
			givenBackendStatus1: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.NatsBackendType,
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionControllerReady, Status: corev1.ConditionTrue},
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionTrue},
				},
				EventingReady: utils.BoolPtr(true),
			},
			givenBackendStatus2: eventingv1alpha1.EventingBackendStatus{
				Backend: eventingv1alpha1.NatsBackendType,
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionControllerReady, Status: corev1.ConditionTrue},
					{Type: eventingv1alpha1.ConditionPublisherProxyReady, Status: corev1.ConditionTrue},
				},
				EventingReady: utils.BoolPtr(true),
			},
			wantResult: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if IsBackendStatusEqual(tc.givenBackendStatus1, tc.givenBackendStatus2) != tc.wantResult {
				t.Errorf("expected output to be %t", tc.wantResult)
			}
		})
	}
}

func Test_isSubscriptionStatusEqual(t *testing.T) {
	testCases := []struct {
		name                string
		subscriptionStatus1 eventingv1alpha2.SubscriptionStatus
		subscriptionStatus2 eventingv1alpha2.SubscriptionStatus
		wantEqualStatus     bool
	}{
		{
			name: "should not be equal if the conditions are not equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionFalse},
				},
				Ready: true,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the ready status is not equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: false,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
				Backend: eventingv1alpha2.Backend{
					APIRuleName: "APIRule",
				},
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
				Backend: eventingv1alpha2.Backend{
					APIRuleName: "APIRule",
				},
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotEqualStatus := IsSubscriptionStatusEqual(tc.subscriptionStatus1, tc.subscriptionStatus2)
			require.Equal(t, tc.wantEqualStatus, gotEqualStatus)
		})
	}
}

func TestPublisherProxyDeploymentEqual(t *testing.T) {
	publisherCfg := env.PublisherConfig{
		Image:          "publisher",
		PortNum:        0,
		MetricsPortNum: 0,
		ServiceAccount: "publisher-sa",
		Replicas:       1,
		RequestsCPU:    "32m",
		RequestsMemory: "64Mi",
		LimitsCPU:      "64m",
		LimitsMemory:   "128Mi",
	}
	natsConfig := env.NATSConfig{
		EventTypePrefix: "prefix",
		JSStreamName:    "kyma",
	}
	defaultNATSPublisher := deployment.NewNATSPublisherDeployment(natsConfig, publisherCfg)
	defaultBEBPublisher := deployment.NewBEBPublisherDeployment(publisherCfg)

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
		"should be unequal if replicas changes": {
			getPublisher1: func() *appsv1.Deployment {
				replicas := int32(2)
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Replicas = &replicas
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				return defaultNATSPublisher.DeepCopy()
			},
			expectedResult: false,
		},
		"should be equal if spec annotations are nil and empty": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = nil
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{}
				return p
			},
			expectedResult: true,
		},
		"should be unequal if spec annotations changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{"key": "value1"}
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Annotations = map[string]string{"key": "value2"}
				return p
			},
			expectedResult: false,
		},
		"should be equal if spec Labels are nil and empty": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = nil
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{}
				return p
			},
			expectedResult: true,
		},
		"should be unequal if spec Labels changes": {
			getPublisher1: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{"key": "value1"}
				return p
			},
			getPublisher2: func() *appsv1.Deployment {
				p := defaultNATSPublisher.DeepCopy()
				p.Spec.Template.Labels = map[string]string{"key": "value2"}
				return p
			},
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if publisherProxyDeploymentEqual(tc.getPublisher1(), tc.getPublisher2()) != tc.expectedResult {
				t.Errorf("expected output to be %t", tc.expectedResult)
			}
		})
	}
}

func Test_ownerReferencesDeepEqual(t *testing.T) {
	ownerReference := func(version, kind, name, uid string, controller, block *bool) metav1.OwnerReference {
		return metav1.OwnerReference{
			APIVersion:         version,
			Kind:               kind,
			Name:               name,
			UID:                types.UID(uid),
			Controller:         controller,
			BlockOwnerDeletion: block,
		}
	}

	tests := []struct {
		name                  string
		givenOwnerReferences1 []metav1.OwnerReference
		givenOwnerReferences2 []metav1.OwnerReference
		wantEqual             bool
	}{
		{
			name:                  "both OwnerReferences are nil",
			givenOwnerReferences1: nil,
			givenOwnerReferences2: nil,
			wantEqual:             true,
		},
		{
			name:                  "both OwnerReferences are empty",
			givenOwnerReferences1: []metav1.OwnerReference{},
			givenOwnerReferences2: []metav1.OwnerReference{},
			wantEqual:             true,
		},
		{
			name: "same OwnerReferences and same order",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-1", "k-1", "n-1", "u-1", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-2", "k-2", "n-2", "u-2", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-1", "k-1", "n-1", "u-1", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-2", "k-2", "n-2", "u-2", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: true,
		},
		{
			name: "same OwnerReferences but different order",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-1", "k-1", "n-1", "u-1", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-2", "k-2", "n-2", "u-2", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-2", "k-2", "n-2", "u-2", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
				ownerReference("v-1", "k-1", "n-1", "u-1", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: true,
		},
		{
			name: "different OwnerReference APIVersion",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-1", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: false,
		},
		{
			name: "different OwnerReference Kind",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-1", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: false,
		},
		{
			name: "different OwnerReference Name",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-1", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: false,
		},
		{
			name: "different OwnerReference UID",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-1", pointer.Bool(false), pointer.Bool(false)),
			},
			wantEqual: false,
		},
		{
			name: "different OwnerReference Controller",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(true), pointer.Bool(false)),
			},
			wantEqual: false,
		},
		{
			name: "different OwnerReference BlockOwnerDeletion",
			givenOwnerReferences1: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(false)),
			},
			givenOwnerReferences2: []metav1.OwnerReference{
				ownerReference("v-0", "k-0", "n-0", "u-0", pointer.Bool(false), pointer.Bool(true)),
			},
			wantEqual: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantEqual, ownerReferencesDeepEqual(tc.givenOwnerReferences1, tc.givenOwnerReferences2))
		})
	}
}
