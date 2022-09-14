//go:build unit
// +build unit

package object_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
)

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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if object.IsBackendStatusEqual(tc.givenBackendStatus1, tc.givenBackendStatus2) != tc.wantResult {
				t.Errorf("expected output to be %t", tc.wantResult)
			}
		})
	}
}

func Test_isSubscriptionStatusEqual(t *testing.T) {
	testCases := []struct {
		name                string
		subscriptionStatus1 eventingv1alpha1.SubscriptionStatus
		subscriptionStatus2 eventingv1alpha1.SubscriptionStatus
		wantEqualStatus     bool
	}{
		{
			name: "should not be equal if the conditions are not equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				},
				Ready: true,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the ready status is not equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: false,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready:       true,
				APIRuleName: "APIRule",
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready:       true,
				APIRuleName: "APIRule",
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if gotEqualStatus := object.IsSubscriptionStatusEqual(tc.subscriptionStatus1, tc.subscriptionStatus2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The Subsciption Status are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}
